package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/misc"
)

func timeoutCtx(seconds int) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	return ctx
}

func (c QueriesX) GetMachines() {
	c.q.GetMachines(context.Background())
}

func (c QueriesX) CreateMachine(machineId misc.MachineId) error {
	return c.q.CreateMachine(context.Background(), string(machineId))
}

func (c QueriesX) DeleteMachine(machineId misc.MachineId) error {
	return c.q.DeleteMachine(context.Background(), string(machineId))
}

func (c QueriesX) GetMonitor() misc.MachineId {
	s, e := c.q.GetMonitor(context.Background())
	if e != nil {
		return misc.NilMachineId
	}
	return misc.MachineId(s)
}

func (c QueriesX) SetMachineAsMonitor(machineId misc.MachineId) error {
	return c.q.SetMachineAsMonitor(timeoutCtx(5), string(machineId))
}

func (c QueriesX) TouchMachine(machineId misc.MachineId) error {
	return c.q.TouchMachine(timeoutCtx(5), string(machineId))
}
