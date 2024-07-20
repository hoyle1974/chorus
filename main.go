package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/charmbracelet/log"

	"github.com/hoyle1974/chorus/connection"
	"github.com/hoyle1974/chorus/room"
)

func main() {
	handler := log.New(os.Stderr)
	logger := slog.New(handler)

	ln, err := net.Listen("tcp", ":8181") // Port can be changed here
	if err != nil {
		logger.Error("Error listening:", err)
		return
	}
	defer ln.Close()

	// This stays the same for now as there is only 1 server
	lobbyId := room.GetGlobalLobby(logger)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		room.RemoveAllRooms()
		os.Exit(0)
	}()

	logger.Info("Server listening on :8181")
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error("Error accepting connection:", err)
			continue
		}
		logger.Info("Client connected:", conn.RemoteAddr())

		c := connection.NewConnection(logger, conn)

		go c.Run(lobbyId)
	}
}
