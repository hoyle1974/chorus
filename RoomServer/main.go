package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"

	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/room"
	"github.com/hoyle1974/chorus/store"
)

var machineId = machine.NewMachineId("RS")

func main() {
	handler := log.New(os.Stderr)
	logger := slog.New(handler)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		// Do any cleanup
		os.Exit(0)
	}()

	// See if we should be the owner of a room
	go func() {
		WaitForOwnership(logger, room.GetGlobalLobbyId(), machineId)

		room := GetRoom(logger, room.GetGlobalLobbyId(), "matchmaker.js")
		logger.Info("Global Lobby", "room", room)
	}()

	logger.Info("RoomServer started.")
	for {
		time.Sleep(time.Second)
	}

}

func WaitForOwnership(logger *slog.Logger, roomId misc.RoomId, machineId misc.MachineId) {
	// See if we can be the owner of this room, block until we can
	key := OwnershipKey(roomId)
	ttl := time.Duration(10) * time.Second

	for {
		ok, _ := store.PutIfAbsent(key, string(machineId), ttl)
		if ok {
			break
		}
		time.Sleep(ttl)
	}
	// We are the owner, spawn a touch point
	go func() {
		for {
			// Touch the record, then sleep for 80% of the ttl
			store.Put(key, string(machineId), ttl)
			time.Sleep(ttl * 8 / 10)
		}
	}()
	// When this function returns we own the room
	logger.Info("We are the owner", "room", roomId)
}

type Room struct {
	logger *slog.Logger
	key    string
	roomId misc.RoomId
}

func (r Room) Destroy() {
	r.logger.Info("Deleting room")
	store.Del(r.key)
}

func GetRoom(logger *slog.Logger, roomId misc.RoomId, adminScript string) Room {
	key := RoomKey(roomId)
	store.Put(key, adminScript, time.Duration(60)*time.Second)

	return Room{key: key,
		roomId: roomId,
		logger: logger.With("roomId", roomId, "script", adminScript),
	}
}
