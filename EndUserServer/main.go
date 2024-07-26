package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/charmbracelet/log"
	"github.com/hoyle1974/chorus/ownership"
)

func main() {
	handler := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
	logger := slog.New(handler)

	state := NewGlobalState(logger)

	ownership.StartLocalOwnershipService(state)

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
