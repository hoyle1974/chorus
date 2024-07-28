package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetConnections() {
	q().GetConnections(context.Background())
}

func GetExpiredConnections() {
	interval := pgtype.Interval{
		Days:         0,
		Months:       0,
		Microseconds: (time.Duration(5) * time.Second).Microseconds(),
		Valid:        true,
	}
	q().GetExpiredConnections(context.Background(), interval)
}

func FindMachine(id misc.ConnectionId) {
	q().FindMachine(context.Background(), string(id))
}

func CreateConnection(connectionId misc.ConnectionId, machineId misc.MachineId) {
	q().CreateConnection(context.Background(), db.CreateConnectionParams{Uuid: string(connectionId), MachineUuid: string(machineId)})
}

func DeleteConnection(connectionId misc.ConnectionId) {
	q().DeleteConnection(context.Background(), string(connectionId))
}
