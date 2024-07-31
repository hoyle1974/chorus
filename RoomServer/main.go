package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/leader"

	"github.com/charmbracelet/log"
)

func onLeaderStartFunc(logger *slog.Logger, q dbx.QueriesX) {
	// logger.Debug("onLeaderStartFunc")
}
func onLeaderTickFunc(logger *slog.Logger, q dbx.QueriesX) {
	// logger.Debug("onLeaderTickFunc")
}

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)
	state := NewGlobalState(logger)

	leader, err := leader.StartLeaderService(state, onLeaderStartFunc, onLeaderTickFunc)
	if err != nil {
		panic(err)
	}

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		// Do any cleanup
		leader.Destroy()

		os.Exit(0)
	}()

	// conn, err := dbx.NewConn()
	// if err != nil {
	// 	panic(err)
	// }
	// q := db.New(conn)
	// m, err := q.GetMachines(context.Background())
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(m)

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
		time.Sleep(time.Second)
	}

}
