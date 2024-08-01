package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

func timeoutCtx(seconds int) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	return ctx
}

func (c QueriesX) GetMachines() {
	c.q.GetMachines(context.Background())
}

func (c QueriesX) GetMachinesByType(machineType string) ([]Machine, error) {
	ms, err := c.q.GetMachinesByType(context.Background(), machineType)
	machines := []Machine{}
	if err != nil {
		return machines, err
	}
	for _, dbMachine := range ms {
		machines = append(machines, toMachine(dbMachine))
	}
	return machines, err
}

func (c QueriesX) CreateMachine(machineId misc.MachineId, machineType string) error {
	return c.q.CreateMachine(context.Background(), db.CreateMachineParams{
		Uuid:        string(machineId),
		MachineType: machineType,
	})
}

func (c QueriesX) DeleteMachine(machineId misc.MachineId) error {
	return c.q.DeleteMachine(context.Background(), string(machineId))
}

type Machine struct {
	Uuid        misc.MachineId
	MachineType string
	CreatedAt   time.Time
	LastUpdated time.Time
}

func toMachine(in db.Machine) Machine {
	return Machine{
		Uuid:        misc.MachineId(in.Uuid),
		MachineType: in.MachineType,
		CreatedAt:   in.CreatedAt.Time,
		LastUpdated: in.LastUpdated.Time,
	}
}

func (c QueriesX) GetMachine(machineId misc.MachineId) (Machine, error) {
	s, err := c.q.GetMachine(context.Background(), string(machineId))
	return toMachine(s), err
}

func (c QueriesX) IsMachineOnline(machineId misc.MachineId) (bool, error) {
	machine, err := c.GetMachine(machineId)
	if err != nil {
		return false, err
	}
	if time.Since(machine.LastUpdated) < time.Duration(5)*time.Second {
		return true, nil
	}
	return false, nil
}

func (c QueriesX) TouchMachine(machineId misc.MachineId) error {
	return c.q.TouchMachine(timeoutCtx(5), string(machineId))
}

func (c QueriesX) GetLeaderForType(machineType string) misc.MachineId {
	s, err := c.q.GetLeaderForType(context.Background(), machineType)
	if err != nil {
		return misc.NilMachineId
	}
	return misc.MachineId(s)
}

func (c QueriesX) SetMachineAsLeader(uuid misc.MachineId) error {
	return c.q.SetMachineAsLeader(context.Background(), string(uuid))
}
