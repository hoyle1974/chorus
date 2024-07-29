package dbx

import (
	"context"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

func (c QueriesX) GetConnections() {
	c.q.GetConnections(context.Background())
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
