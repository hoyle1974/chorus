package main

import (
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/store"
)

type GlobalServerState struct {
	logger       *slog.Logger
	MachineId    misc.MachineId
	MachineLease *store.Lease
}

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:       logger,
		MachineId:    machine.NewMachineId("RS"),
		MachineLease: store.NewLease(time.Duration(10) * time.Second),
	}

	err := store.Put(ss.MachineId.MachineKey(), "true", ss.MachineLease.TTL)
	if err != nil {
		panic(err)
	}
	ss.MachineLease.AddKey(ss.MachineId.MachineKey())

	return ss
}
