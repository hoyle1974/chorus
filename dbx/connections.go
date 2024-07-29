package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
	"github.com/jackc/pgx/v5/pgtype"
)

func (c QueriesX) GetConnections() {
	c.q.GetConnections(context.Background())
}

func (c QueriesX) GetExpiredConnections() {
	interval := pgtype.Interval{
		Days:         0,
		Months:       0,
		Microseconds: (time.Duration(5) * time.Second).Microseconds(),
		Valid:        true,
	}
	c.q.GetExpiredConnections(context.Background(), interval)
}

func (c QueriesX) FindMachine(id misc.ConnectionId) {
	c.q.FindMachine(context.Background(), string(id))
}

func (c QueriesX) CreateConnection(connectionId misc.ConnectionId, machineId misc.MachineId) {
	c.q.CreateConnection(context.Background(), db.CreateConnectionParams{Uuid: string(connectionId), MachineUuid: string(machineId)})
}

func (c QueriesX) DeleteConnection(connectionId misc.ConnectionId) {
	c.q.DeleteConnection(context.Background(), string(connectionId))
}
