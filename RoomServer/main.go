package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/ownership"
)

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)
	state := NewGlobalState(logger)

	state.ownership = ownership.StartLocalOwnershipService(state)
	rs := StartLocalRoomService(state)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		// Do any cleanup
		state.machineLease.Destroy()
		state.ownership.StopLocalService()
		rs.StopLocalService()

		os.Exit(0)
	}()

	// See if we should be the owner of a room
	if state.ownership.GetValidOwner(misc.GetGlobalLobbyId()) == misc.NilMachineId {
		rs.BootstrapLobby()
	}

	rs.gss.logger.Info("RoomServer started.")
	for {
		time.Sleep(time.Second)
	}

}
