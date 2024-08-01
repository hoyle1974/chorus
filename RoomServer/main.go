package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/hoyle1974/chorus/leader"
	"github.com/hoyle1974/chorus/misc"

	"github.com/charmbracelet/log"
)

func onLeaderStartFunc(ctx leader.LeaderQueryContext) {
	// logger.Debug("onLeaderStartFunc")
}
func onLeaderTickFunc(ctx leader.LeaderQueryContext) {
	// logger.Debug("onLeaderTickFunc")
}
func onMachineOffline(ctx leader.LeaderQueryContext, machineId misc.MachineId) {
	// We have a machine that is offline, what cleanup should we do?
	ctx.Logger().Info("onMachineOffline", "offlineMachine", machineId)
	rooms, err := ctx.Query().GetRoomsByMachine(machineId)
	if err != nil {
		ctx.Logger().Error("Could not get rooms by machine", "error", err)
		return
	}
	for _, room := range rooms {
		if room.DestroyOnOrphan {
			ctx.Query().DeleteRoom(room.Uuid)
		} else {
			// Someone needs to own this, for now it's us
			err = ctx.Query().SetRoomOwner(room.Uuid, machineId, ctx.MachineId())
			if err == nil {
				// Successful, bind it locally
				rs.bindRoomToThisMachine(RoomInfo{
					RoomId:          room.Uuid,
					AdminScript:     room.Script,
					Name:            room.Name,
					DestroyOnOrphan: room.DestroyOnOrphan,
				})
			} else {
				ctx.Logger().Error("Error becoming owner of room", "room", room, "error", err)
			}
		}
	}
}

var rs *RoomService

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)
	state := NewGlobalState(logger)

	leader, err := leader.StartLeaderService(state, onLeaderStartFunc, onLeaderTickFunc, onMachineOffline)
	if err != nil {
		panic(err)
	}

	rs = StartLocalRoomService(state)

	if rs.BootstrapLobby() {
		state.logger.Info("Global Lobby boostrapped")
	}

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		// Do any cleanup
		leader.Destroy()

		os.Exit(0)
	}()

	/*
		state.ownership = ownership.StartLocalOwnershipService(state)
		rs := StartLocalRoomService(state)



		// See if we should be the owner of a room
		if state.ownership.GetValidOwner(misc.GetGlobalLobbyId()) == misc.NilMachineId {
			rs.BootstrapLobby()
		}
	*/

	state.logger.Info("RoomServer started.")
	for {
		time.Sleep(time.Hour)
	}

}
