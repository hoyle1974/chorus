package monitor

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
 * one machine in the cluster is the monitor
 *
 * Monitor - continually scans the monitor table.  Machines that are too old, get deleted.
 *
 * Not Monitor - watches the monitor to make sure it updates the table,
 * if it does not, it tries ot become the monitor
 */

type MonitorContext interface {
	Logger() *slog.Logger
	MachineId() misc.MachineId
}

type MonitorService struct {
	dbx       dbx.DBX
	machineId misc.MachineId
	logger    *slog.Logger
}

func (ms MonitorService) Destroy() error {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	err := q.DeleteMachine(ms.machineId)
	return err
}

func StartMonitorService(ctx MonitorContext) (MonitorService, error) {
	ms := MonitorService{
		logger:    ctx.Logger().With("machineId", ctx.MachineId()),
		machineId: ctx.MachineId(),
		dbx:       dbx.Dbx(),
	}
	defer ms.logger.Info("Monitor Service Started . . .")

	// Start transaction
	tx, err := dbx.GetConn().Begin(context.Background())
	if err != nil {
		return ms, err
	}
	defer tx.Rollback(context.Background())

	// Create ourselves as a machine in the table
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()).WithTx(tx))
	err = q.CreateMachine(ctx.MachineId())
	if err != nil {
		return ms, err
	}
	go ms.keepAliveTick()

	monitorMachineId := q.GetMonitor()
	if monitorMachineId == misc.NilMachineId {
		// Let's become the monitor
		err := q.SetMachineAsMonitor(ctx.MachineId())
		if err != nil {
			// Nope we are not the monitor
			go ms.waitForMonitor()
			return ms, nil
		}
		// We are the monitor
		ms.becomeMonitor(q)
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return ms, err
	}

	go ms.waitForMonitor()
	return ms, nil
}

func (ms MonitorService) keepAliveTick() {
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

func (ms MonitorService) becomeMonitor(q dbx.QueriesX) error {
	ms.logger.Debug("becomeMonitor")

	err := q.SetMachineAsMonitor(ms.machineId)
	if err == nil {
		ms.logger.Info("We are the monitor")
		ms.logger = ms.logger.With("monitor", true)
		go ms.monitor()
	} else {
		ms.logger.Error("Not able to become monitor at the moment.")
	}

	return err
}

func (ms MonitorService) monitor() {
	ms.logger.Debug("monitor")

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

// We are not the monitor, but wait to see if we can become a monitor
func (ms MonitorService) waitForMonitor() {
	ms.logger.Debug("waitForMonitor")

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

			// Timeout occured, the monitor is no longer valid, let's become the monitor
			machineId := q.GetMonitor()
			if machineId != misc.NilMachineId {
				machine, err := q.GetMachine(machineId)
				if err == nil {
					// Monitor machine has timed out
					if time.Now().Sub(machine.LastUpdated).Seconds() > 5 {
						ms.logger.Warn("Monitor deadline exceeded, let's try to become the monitor now")

						err = q.DeleteMachine(machineId)
						if err == nil {
							err = ms.becomeMonitor(q)
							if err == nil {
								// We are the monitor, no more waiting
								tx.Commit(context.Background())
								return
							} else {
								ms.logger.Error("could not become the monitor", "error", err)
							}
						} else {
							ms.logger.Error("could not delete expired monitor", "error", err)
						}
					}
				}
			} else {
				ms.logger.Error("could not get a monitor", "error", err)
			}
			tx.Rollback(context.Background())
		} else {
			ms.logger.Error("Error waiting for notification", "error", err)
			panic(err)
		}
	}

}
