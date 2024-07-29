package monitor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/misc"
	"github.com/jackc/pgx/v5/pgtype"
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
	ctx MonitorContext
	dbx dbx.DBX
}

func StartMonitorService(ctx MonitorContext) MonitorService {
	ms := MonitorService{
		ctx: ctx,
		dbx: dbx.Dbx(),
	}

	// Start transaction
	tx, err := dbx.GetConn().Begin(context.Background())
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(context.Background())

	// Create ourselves as a machine in the table
	q := db.New(dbx.GetConn()).WithTx(tx)
	err = ms.dbx.Queries(q).CreateMachine(ctx.MachineId())
	if err != nil {
		panic(err)
	}

	monitorMachineId := ms.dbx.Queries(q).GetMonitor()
	if monitorMachineId == misc.NilMachineId {
		// Let's become the monitor
		err := ms.dbx.Queries(q).SetMachineAsMonitor(ctx.MachineId())
		if err != nil {
			// Nope we are not the monitor
			go ms.waitForMonitor()
			return ms
		}
		// We are the monitor
		ms.becomeMonitor(q)
	}
	tx.Commit(context.Background())

	go ms.waitForMonitor()
	return ms
}

func (ms MonitorService) becomeMonitor(q *db.Queries) error {
	err := ms.dbx.Queries(q).SetMachineAsMonitor(ms.ctx.MachineId())
	if err == nil {
		fmt.Printf("We are the monitor: [%v]\n", ms.ctx.MachineId())

		go ms.monitor()
	} else {
		fmt.Printf("Not able to become monitor at the moment.")
	}

	return err
}

func (ms MonitorService) monitor() {
	for {
		fmt.Println("monitor tick")
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		time.Sleep(time.Duration(1) * time.Second)
		conn, err := dbx.NewConn()
		if err == nil {
			fmt.Println("expired machines?")
			q := db.New(conn)
			uuids, err := q.GetExpiredMachines(ctx, pgtype.Interval{
				Microseconds: (time.Duration(3) * time.Second).Microseconds(),
				Days:         0,
				Months:       0,
				Valid:        true,
			})
			if err != nil {
				for _, uuid := range uuids {
					fmt.Println("Delete machine", uuid)
					q.DeleteMachine(ctx, uuid)
				}
			} else {
				fmt.Println(err)
			}

			fmt.Println("expired connections?")
			uuids, err = q.GetExpiredConnections(ctx, pgtype.Interval{
				Microseconds: (time.Duration(3) * time.Second).Microseconds(),
				Days:         0,
				Months:       0,
				Valid:        true,
			})
			if err != nil {
				for _, uuid := range uuids {
					fmt.Println("Delete Connection", uuid)
					q.DeleteConnection(ctx, uuid)
				}
			} else {
				fmt.Println(err)
			}
		}
	}
}

func (ms MonitorService) waitForMonitor() {
	conn, err := dbx.NewConn()
	if err != nil {
		panic(err)
	}

	// Listen to a channel named "my_channel"
	_, err = conn.Exec(context.Background(), "LISTEN machines")
	if err != nil {
		panic(err)
	}

	fmt.Println("-- Start to listen")
	for {
		fmt.Println("Wait for notification")
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("Monitor deadline exceeded, let's try to become the monitor now")

				// Start transaction
				tx, err := dbx.GetConn().Begin(context.Background())
				if err != nil {
					panic(err)
				}
				q := db.New(dbx.GetConn())

				// Timeout occured, the monitor is no longer valid, let's become the monitor
				machineId, err := q.GetMonitor(context.Background())
				if err == nil {
					err = q.DeleteMachine(context.Background(), machineId)
					if err == nil {
						err = ms.becomeMonitor(q)
						if err == nil {
							// We are the monitor, no more waiting
							tx.Commit(context.Background())
							return
						} else {
							fmt.Println(err)
						}
					} else {
						fmt.Println(err)
					}
				} else {
					fmt.Println(err)
				}
				tx.Rollback(context.Background())
			} else {
				fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
				os.Exit(1)
			}
		} else {
			if notification.Channel == "machines" && notification.Payload == "monitor data updated" {
				// See if we need to become the monitor?
			}
		}
	}

}
