package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/leader"
	"github.com/hoyle1974/chorus/ownership"
)

func onLeaderStartFunc(logger *slog.Logger, q dbx.QueriesX) {
	// logger.Debug("onLeaderStartFunc")
}
func onLeaderTickFunc(logger *slog.Logger, q dbx.QueriesX) {
	// logger.Debug("onLeaderTickFunc")

	// Cleanup old connections
	connections, err := q.GetConnections()
	now := time.Now()
	if err == nil {
		for _, connection := range connections {
			if now.Sub(connection.LastUpdated).Seconds() > 5 {
				logger.Debug("Delete connection", "connectionId", connection.Uuid)
				err := q.DeleteConnection(connection.Uuid)
				if err != nil {
					logger.Error("Problem deleting connection", "error", err)
				}
			}
		}
	} else {
		logger.Error("Trouble getting a list of all connections", "error", err)
	}
}

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)

	state := NewGlobalState(logger)

	leader, err := leader.StartLeaderService(state, onLeaderStartFunc, onLeaderTickFunc)
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
