package main

import (
	"log/slog"

	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
)

type GlobalServerState struct {
	logger    *slog.Logger
	MachineId misc.MachineId
}

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:    logger,
		MachineId: machine.NewMachineId("EUS"),
	}

	return ss
}
