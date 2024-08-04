package main

import (
	"log/slog"

	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
)

type GlobalServerState struct {
	logger    *slog.Logger
	machineId misc.MachineId
}

func (gs GlobalServerState) Logger() *slog.Logger      { return gs.logger }
func (gs GlobalServerState) MachineId() misc.MachineId { return gs.machineId }
func (gs GlobalServerState) MachineType() string       { return "RoomServer" }

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:    logger,
		machineId: machine.NewMachineId("RS"),
	}

	return ss
}
