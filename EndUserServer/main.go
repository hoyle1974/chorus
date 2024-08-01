package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hoyle1974/chorus/leader"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/ownership"
)

func onLeaderStartFunc(ctx leader.LeaderQueryContext) {
	// logger.Debug("onLeaderStartFunc")
}
func onLeaderTickFunc(ctx leader.LeaderQueryContext) {
	// logger.Debug("onLeaderTickFunc")

	// Cleanup old connections
	connections, err := ctx.Query().GetConnections()
	now := time.Now()
	if err == nil {
		for _, connection := range connections {
			if now.Sub(connection.LastUpdated).Seconds() > 5 {
				ctx.Logger().Debug("Delete connection", "connectionId", connection.Uuid)
				err := ctx.Query().DeleteConnection(connection.Uuid)
				if err != nil {
					ctx.Logger().Error("Problem deleting connection", "error", err)
				}
			}
		}
	} else {
		ctx.Logger().Error("Trouble getting a list of all connections", "error", err)
	}
}
func onMachineOffline(ctx leader.LeaderQueryContext, machineId misc.MachineId) {
	// We have a machine that is offline, what cleanup should we do?

}

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)

	state := NewGlobalState(logger)

	leader, err := leader.StartLeaderService(state, onLeaderStartFunc, onLeaderTickFunc, onMachineOffline)
	if err != nil {
		panic(err)
	}

	state.ownership = ownership.StartLocalOwnershipService(state)

	ln, err := net.Listen("tcp", ":8181") // Port can be changed here
	if err != nil {
		logger.Error("Error listening:", err)
		return
	}
	defer ln.Close()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		// Do any cleanup
		leader.Destroy()
		cleanupConnections()
		os.Exit(0)
	}()

	state.logger.Info("EndUserServer listening on :8181")
	for {
		conn, err := ln.Accept()
		if err != nil {
			state.logger.Error("Error accepting connection:", err)
			continue
		}
		state.logger.Info("Client connected:", conn.RemoteAddr())

		c := NewConnection(state, conn)
		go c.Run()
	}
}
