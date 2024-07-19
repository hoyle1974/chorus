package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/charmbracelet/log"
)

func main() {
	handler := log.New(os.Stderr)
	logger := slog.New(handler)

	ln, err := net.Listen("tcp", ":8080") // Port can be changed here
	if err != nil {
		logger.Error("Error listening:", err)
		return
	}
	defer ln.Close()

	room, err := NewRoom(logger, "Default Lobby", "matchmaker.js")
	if err != nil {
		logger.Error("Error creating default lobby", err)
		return
	}

	logger.Info("Server listening on :8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error("Error accepting connection:", err)
			continue
		}
		logger.Info("Client connected:", conn.RemoteAddr())

		c := NewConnection(logger, conn)

		go c.Run(room)
	}
}
