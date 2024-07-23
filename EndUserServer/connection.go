package main

import (
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"github.com/hoyle1974/chorus/room"
	"github.com/hoyle1974/chorus/store"
)

type Connection struct {
	logger   *slog.Logger
	id       misc.ConnectionId
	conn     net.Conn
	consumer *pubsub.Consumer
}

var connectionLock sync.Mutex
var connections = map[misc.ConnectionId]*Connection{}

func cleanupConnections() {
	connectionLock.Lock()
	defer connectionLock.Unlock()
	for _, v := range connections {
		v.Close()
	}
}

// func FindConnectionById(id misc.ConnectionId) *Connection {
// 	connectionLock.Lock()
// 	defer connectionLock.Unlock()

// 	return connections[id]
// }

// // OnMessage implements RoomListener.
// func (c *Connection) OnMessage(msg message.Message) {
// 	if c.conn == nil {
// 		return
// 	}
// 	c.conn.Write([]byte(msg.String() + "\n"))
// }

func NewConnection(logger *slog.Logger, conn net.Conn) *Connection {
	c := Connection{
		id:   misc.ConnectionId("C" + misc.UUIDString()),
		conn: conn,
	}
	c.logger = logger.With("connectionId", c.id)

	store.PutConnectionInfo(machineId, c.id)

	connectionLock.Lock()
	connections[c.id] = &c
	connectionLock.Unlock()

	return &c
}

func (c *Connection) Close() {
	c.logger.Info("Closing connection")

	room.LeaveAllRooms(misc.ListenerId(c.id))
	store.RemoveConnectionInfo(machineId, c.id)

	if c.conn != nil {
		c.conn.Close()
	}

	connectionLock.Lock()
	delete(connections, c.id)
	connectionLock.Unlock()

}

func (c *Connection) OnMessageFromTopic(msg message.Message) {
	c.logger.Info("Connection.OnMessageFromTopic", "msg", msg)
	if c.conn != nil && (msg.ReceiverId == misc.ListenerId(c.id) || msg.ReceiverId == "") {
		c.conn.Write([]byte(msg.String() + "\n"))
	}
}

func (c *Connection) Run() {
	defer c.Close()

	// We have a new connection, let's join the global lobby
	room.Join(room.GetGlobalLobbyId(), misc.ListenerId(c.id), c)

	buf := make([]byte, 65536)
	for {
		n, err := c.conn.Read(buf)
		if err == io.EOF {
			c.logger.Info("Disconnect user")
			c.conn = nil
			return
		}
		if err != nil {
			c.logger.Error("connection error", err)
			c.conn = nil
			return
		}
		// Trim trailing newline (if present)
		words := strings.Fields(strings.TrimSpace(string(buf[:n])))
		if len(words) == 1 {
			if words[0] == "exit" {
				return
			}
		}

		if len(words) < 2 {
			continue
		}
		msg := message.NewMessage(misc.RoomId(words[0]), misc.ListenerId(c.id), "room", string(words[1]), map[string]interface{}{})

		for t := 0; t < len(words)-2; t += 2 {
			key := words[t+2]
			value := words[t+3]
			msg.Data[key] = value
		}
		room.Send(msg)
	}

}
