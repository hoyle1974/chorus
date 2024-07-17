package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Connection struct {
	id   ListenerId
	conn net.Conn
}

var connectionLock sync.Mutex
var connections = map[ListenerId]*Connection{}

func FindConnectionById(id ListenerId) *Connection {
	connectionLock.Lock()
	defer connectionLock.Unlock()

	return connections[id]
}

// OnMessage implements RoomListener.
func (c *Connection) OnMessage(msg Message) {
	fmt.Println("OnMessage: ", msg)
	c.conn.Write([]byte(msg.String() + "\n"))
}

func NewConnection(conn net.Conn) *Connection {
	c := Connection{
		id:   ListenerId("L" + uuid.NewString()),
		conn: conn,
	}

	connectionLock.Lock()
	defer connectionLock.Unlock()
	connections[c.id] = &c

	return &c
}

func (c *Connection) Run(room *Room) {
	defer c.conn.Close()

	// We have a new connection, let's join the global lobby

	msg := NewMessage(room.id, c.id, "", "init", map[string]string{})
	c.conn.Write([]byte(msg.String() + "\n"))

	room.Join(c.id, c)

	buf := make([]byte, 65536)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			c.conn.Write([]byte(NewErrorMessage(room.id, c.id, err).String() + "\n"))
			return
		}

		// Trim trailing newline (if present)
		words := strings.Fields(strings.TrimSpace(string(buf[:n])))

		msg := NewMessage(RoomId(words[0]), c.id, "room", string(words[1]), map[string]string{})

		for t := 0; t < len(words)-2; t += 2 {
			key := words[t+2]
			value := words[t+3]
			msg.Data[key] = value
		}

		FindRoom(msg.RoomId).Send(msg)
	}

}
