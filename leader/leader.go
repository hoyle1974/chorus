package leader

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/misc"
)

/*
 * one machine in the cluster is the leader per machine type
 *
 * leader - continually scans the leader & machines table.  Machines that are too old, get deleted.
 *
 * Not leader - watches the leader to make sure it updates the table,
 * if it does not, it tries ot become the leader
 */

type LeaderContext interface {
	Logger() *slog.Logger
	MachineId() misc.MachineId
	MachineType() string
}

type LeaderService struct {
	dbx              dbx.DBX
	machineId        misc.MachineId
	logger           *slog.Logger
	machineType      string
	onLeaderStart    onLeader
	onLeaderTick     onLeader
	onMachineOffline onMachineOffline
}

func (ms LeaderService) Destroy() error {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	err := q.DeleteMachine(ms.machineId)
	return err
}

type LeaderQueryContext interface {
	LeaderContext
	Query() dbx.QueriesX
}

type leaderQueryContextImpl struct {
	logger      *slog.Logger
	machineId   misc.MachineId
	machineType string
	q           dbx.QueriesX
}

func (l leaderQueryContextImpl) Logger() *slog.Logger      { return l.logger }
func (l leaderQueryContextImpl) MachineId() misc.MachineId { return l.machineId }
func (l leaderQueryContextImpl) MachineType() string       { return l.machineType }
func (l leaderQueryContextImpl) Query() dbx.QueriesX       { return l.q }

type onLeader func(ctx LeaderQueryContext)
type onMachineOffline func(ctx LeaderQueryContext, machineId misc.MachineId)

func StartLeaderService(ctx LeaderContext, onLeaderStart onLeader, onLeaderTick onLeader, onMachineOffline onMachineOffline) (LeaderService, error) {
	ms := LeaderService{
		logger:           ctx.Logger().With("machineId", ctx.MachineId(), "type", ctx.MachineType()),
		machineId:        ctx.MachineId(),
		dbx:              dbx.Dbx(),
		machineType:      ctx.MachineType(),
		onLeaderStart:    onLeaderStart,
		onLeaderTick:     onLeaderTick,
		onMachineOffline: onMachineOffline,
	}
	defer ms.logger.Info("Leader Service Started . . .")

	// Create ourselves as a machine in the table
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()) /*.WithTx(tx)*/)
	err := q.CreateMachine(ctx.MachineId(), ctx.MachineType())
	if err != nil {
		return ms, err
	}
	go ms.keepAliveTick()

	// Start transaction
	tx, err := dbx.GetConn().Begin(context.Background())
	if err != nil {
		return ms, err
	}
	defer tx.Rollback(context.Background())

	q = dbx.Dbx().Queries(db.New(dbx.GetConn()).WithTx(tx))

	leaderMachineId := q.GetLeaderForType(ctx.MachineType())
	if leaderMachineId == misc.NilMachineId {
		// Let's become the leader
		err := q.SetMachineAsLeader(ctx.MachineId())
		if err == nil {
			// We are the leader
			ms.becomeLeader(q)
			err = tx.Commit(context.Background())
			return ms, err
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return ms, err
	}

	go ms.waitForLeader()
	return ms, nil
}

func (ms LeaderService) keepAliveTick() {
	ms.logger.Debug("keepAliveTick")
	conn, err := dbx.NewConn()
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
	q := db.New(conn)

	for {
		err := ms.dbx.Queries(q).TouchMachine(ms.machineId)
		if err != nil {
			ms.logger.Error("Could not touch our record in the database", "error", err)
		}
		time.Sleep(time.Second * time.Duration(1))
	}
}

func (ms LeaderService) becomeLeader(q dbx.QueriesX) {
	ms.logger.Info("We are the leader")
	ms.logger = ms.logger.With("leader", true)
	go ms.monitorLeadership()
}

func (ms LeaderService) monitorLeadership() {
	ms.logger.Debug("leader")

	conn, err := dbx.NewConn()
	if err != nil {
		panic(err)
	}
	q := dbx.Dbx().Queries(db.New(conn))

	lqc := leaderQueryContextImpl{
		logger:      ms.logger,
		machineId:   ms.machineId,
		machineType: ms.machineType,
		q:           q,
	}
	ms.onLeaderStart(lqc)

	for {
		time.Sleep(time.Duration(1) * time.Second)
		machines, err := q.GetMachinesByType(ms.machineType)
		now := time.Now()
		if err == nil {
			for _, machine := range machines {
				if now.Sub(machine.LastUpdated).Seconds() > 5 {
					ms.logger.Debug("Delete machine", "machineToDelete", machine.Uuid)
					ms.onMachineOffline(lqc, machine.Uuid)
					err := q.DeleteMachine(machine.Uuid)
					if err != nil {
						ms.logger.Error("Problem deleting machine", "error", err)
					}
				}
			}
		} else {
			ms.logger.Error("Trouble getting a list of all machines", "error", err)
		}

		ms.onLeaderTick(lqc)

	}
}

// We are not the leader, but wait to see if we can become a leader
func (ms LeaderService) waitForLeader() {
	ms.logger.Debug("waitForLeader")

	conn, err := dbx.NewConn()
	if err != nil {
		panic(err)
	}

	// Listen to a channel named "my_channel"
	_, err = conn.Exec(context.Background(), "LISTEN machines")
	if err != nil {
		panic(err)
	}

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		_, err := conn.WaitForNotification(ctx)
		if err == nil || errors.Is(err, context.DeadlineExceeded) {

			// Start transaction
			tx, err := dbx.GetConn().Begin(context.Background())
			if err != nil {
				ms.logger.Error("Could not get connection", "error", err)
			}
			conn := dbx.GetConn()
			dbq := db.New(conn)
			q := dbx.Dbx().Queries(dbq)

			// Timeout occured, the monleaderitor is no longer valid, let's become the leader
			machineId := q.GetLeaderForType(ms.machineType)
			if machineId != misc.NilMachineId {
				machine, err := q.GetMachine(machineId)
				if err == nil {
					// Leader machine has timed out
					if time.Since(machine.LastUpdated).Seconds() > 5 {
						ms.logger.Warn("Leader deadline exceeded, let's try to become the leader now")

						lqc := leaderQueryContextImpl{
							logger:      ms.logger,
							machineId:   ms.machineId,
							machineType: ms.machineType,
							q:           q,
						}
						ms.onMachineOffline(lqc, machine.Uuid)
						err = q.DeleteMachine(machineId)
						if err == nil {
							err := q.SetMachineAsLeader(ms.machineId)
							if err == nil {
								ms.becomeLeader(q)
								err = tx.Commit(context.Background())
								if err != nil {
									panic(err)
								}
								return
							} else {
								ms.logger.Error("could not become the leader", "error", err)
							}
						} else {
							ms.logger.Error("could not delete expired leader", "error", err)
						}
					}
				}
			} else {
				ms.logger.Error("could not get a leader", "error", err)
			}
			tx.Rollback(context.Background())
		} else {
			ms.logger.Error("Error waiting for notification", "error", err)
			panic(err)
		}
	}

}
