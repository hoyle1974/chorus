package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/misc"
	"github.com/jackc/pgx/v5/pgtype"
)

func timeoutCtx(seconds int) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	return ctx
}

func (c QueriesX) GetMachines() {
	c.q.GetMachines(context.Background())
}

func (c QueriesX) GetExpiredMachines() {
	interval := pgtype.Interval{
		Days:         0,
		Months:       0,
		Microseconds: (time.Duration(5) * time.Second).Microseconds(),
		Valid:        true,
	}
	c.q.GetExpiredMachines(context.Background(), interval)
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
