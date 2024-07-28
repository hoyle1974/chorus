package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/misc"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetMachines() {
	q().GetMachines(context.Background())
}

func GetExpiredMachines() {
	interval := pgtype.Interval{
		Days:         0,
		Months:       0,
		Microseconds: (time.Duration(5) * time.Second).Microseconds(),
		Valid:        true,
	}
	q().GetExpiredMachines(context.Background(), interval)
}

func CreateMachine(machineId misc.MachineId) {
	q().CreateMachine(context.Background(), string(machineId))
}

func DeleteMachine(machineId misc.MachineId) {
	q().DeleteMachine(context.Background(), string(machineId))
}
