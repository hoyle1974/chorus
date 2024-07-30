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
	dbx         dbx.DBX
	machineId   misc.MachineId
	logger      *slog.Logger
	machineType string
}

func (ms LeaderService) Destroy() error {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	err := q.DeleteMachine(ms.machineId)
	return err
}

func StartLeaderService(ctx LeaderContext) (LeaderService, error) {
	ms := LeaderService{
		logger:      ctx.Logger().With("machineId", ctx.MachineId()),
		machineId:   ctx.MachineId(),
		dbx:         dbx.Dbx(),
		machineType: ctx.MachineType(),
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

	ttl := time.Duration(5) * time.Second

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		time.Sleep(time.Duration(1) * time.Second)
		conn, err := dbx.NewConn()
		if err == nil {
			q := db.New(conn)
			ctx, _ = context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
			machines, err := q.GetMachines(ctx)
			now := time.Now()
			if err == nil {
				for _, machine := range machines {
					if now.Sub(machine.LastUpdated.Time).Seconds() > 5 {
						ms.logger.Debug("Delete machine", "machineToDelete", machine.Uuid)
						ctx, _ = context.WithTimeout(context.Background(), ttl)
						q.DeleteMachine(ctx, machine.Uuid)
					}
				}
			} else {
				ms.logger.Error("Trouble getting a list of all machines", "error", err)
			}

			ctx, _ = context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
			connections, err := q.GetConnections(ctx)
			if err == nil {
				for _, connection := range connections {
					if now.Sub(connection.LastUpdated.Time).Seconds() > 5 {
						ms.logger.Debug("Delete connection", "connectionToDelete", connection.Uuid)
						ctx, _ = context.WithTimeout(context.Background(), ttl)
						q.DeleteConnection(ctx, connection.Uuid)
					}
				}
			} else {
				ms.logger.Error("Trouble getting a list of all connections", "err", err)
			}
		}
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
					if time.Now().Sub(machine.LastUpdated).Seconds() > 5 {
						ms.logger.Warn("Leader deadline exceeded, let's try to become the leader now")

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
