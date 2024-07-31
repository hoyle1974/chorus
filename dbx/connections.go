package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

type Connection struct {
	Uuid        misc.ConnectionId
	MachineUuid misc.MachineId
	CreatedAt   time.Time
	LastUpdated time.Time
}

func toConnection(in db.Connection) Connection {
	return Connection{
		Uuid:        misc.ConnectionId(in.Uuid),
		MachineUuid: misc.MachineId(in.MachineUuid),
		CreatedAt:   in.CreatedAt.Time,
		LastUpdated: in.LastUpdated.Time,
	}
}

func (c QueriesX) GetConnections() ([]Connection, error) {
	rows, err := c.q.GetConnections(context.Background())
	connections := []Connection{}
	if err != nil {
		return connections, err
	}
	for _, dbConnection := range rows {
		connections = append(connections, toConnection(dbConnection))
	}
	return connections, err
}

func (c QueriesX) FindMachine(id misc.ConnectionId) {
	c.q.FindMachine(context.Background(), string(id))
}

func (c QueriesX) CreateConnection(connectionId misc.ConnectionId, machineId misc.MachineId) {
	c.q.CreateConnection(context.Background(), db.CreateConnectionParams{Uuid: string(connectionId), MachineUuid: string(machineId)})
}

func (c QueriesX) DeleteConnection(connectionId misc.ConnectionId) error {
	return c.q.DeleteConnection(context.Background(), string(connectionId))
}
