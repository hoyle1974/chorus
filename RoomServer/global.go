package main

import (
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/ds"
	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/ownership"
	"github.com/hoyle1974/chorus/store"
)

type GlobalServerState struct {
	logger       *slog.Logger
	machineId    misc.MachineId
	machineLease *store.Lease
	dist         distributed.Dist
	ownership    *ownership.OwnershipService
}

func (gs GlobalServerState) Logger() *slog.Logger      { return gs.logger }
func (gs GlobalServerState) MachineId() misc.MachineId { return gs.machineId }
func (gs GlobalServerState) Dist() distributed.Dist    { return gs.dist }
func (gs GlobalServerState) MachineType() string       { return "RoomServer" }

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:       logger,
		machineId:    machine.NewMachineId("RS"),
		machineLease: store.NewLease(time.Duration(10) * time.Second),
		dist:         distributed.NewDist(ds.GetConn()),
	}
	ss.dist.Put(ss.machineId.MachineKey(), "true", ss.machineLease)

	return ss
}
