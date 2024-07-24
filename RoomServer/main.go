package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"

	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/store"
)

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)
	state := NewGlobalState(logger)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		// Do any cleanup
		os.Exit(0)
	}()

	// See if we should be the owner of a room
	go func() {
		WaitForOwnership(state, misc.GetGlobalLobbyId(), state.MachineId)

		room := GetRoom(state, misc.GetGlobalLobbyId(), "Global Lobby", "matchmaker.js")
		logger.Info("Global Lobby", "room", room)
	}()

	logger.Info("RoomServer started.")
	for {
		time.Sleep(time.Second)
	}

}

func WaitForOwnership(state GlobalServerState, roomId misc.RoomId, machineId misc.MachineId) {
	// See if we can be the owner of this room, block until we can
	ttl := time.Duration(10) * time.Second

	for {
		ok, _ := store.PutIfAbsent(roomId.OwnershipKey(), string(machineId), ttl)
		if ok {
			break
		}
		time.Sleep(ttl)
	}
	// We are the owner, spawn a touch point
	go func() {
		for {
			// Touch the record, then sleep for 80% of the ttl
			store.Put(roomId.OwnershipKey(), string(machineId), ttl)
			time.Sleep(ttl * 8 / 10)
		}
	}()
	// When this function returns we own the room
	state.logger.Info("We are the owner", "room", roomId)
}
