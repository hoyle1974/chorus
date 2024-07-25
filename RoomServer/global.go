package main

import (
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/ds"
	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/store"
)

type GlobalServerState struct {
	logger       *slog.Logger
	MachineId    misc.MachineId
	MachineLease *store.Lease
	Dist         distributed.Dist
}

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:       logger,
		MachineId:    machine.NewMachineId("RS"),
		MachineLease: store.NewLease(time.Duration(10) * time.Second),
		Dist:         distributed.NewDist(ds.GetConn()),
	}
	ss.Dist.Put(ss.MachineId.MachineKey(), "true", ss.MachineLease)

	return ss
}
