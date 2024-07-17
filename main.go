package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080") // Port can be changed here
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()

	room, err := NewRoom("Default Lobby", "matchmaker.js")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	fmt.Println("Server listening on :8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("Client connected:", conn.RemoteAddr())

		c := NewConnection(conn)

		go c.Run(room)
	}
}
