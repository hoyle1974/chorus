package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"github.com/hoyle1974/chorus/store"
)

type Connection struct {
	logger   *slog.Logger
	id       misc.ConnectionId
	conn     net.Conn
	consumer *pubsub.Consumer
	state    GlobalServerState
}

var connectionLock sync.Mutex
var connections = map[misc.ConnectionId]*Connection{}

func cleanupConnections() {
	connectionLock.Lock()
	defer connectionLock.Unlock()
	for _, v := range connections {
		v.Close()
	}
	connections = map[misc.ConnectionId]*Connection{}
}
func cleanupConnection(cid misc.ConnectionId) {
	connectionLock.Lock()
	defer connectionLock.Unlock()
	v, ok := connections[cid]
	if ok {
		v.Close()
	}
	delete(connections, cid)
}

func NewConnection(state GlobalServerState, conn net.Conn) *Connection {
	c := Connection{
		id:    misc.ConnectionId("C" + misc.UUIDString()),
		conn:  conn,
		state: state,
	}
	c.logger = state.logger.With("connectionId", c.id)

	store.PutConnectionInfo(state.MachineId, c.id)

	connectionLock.Lock()
	connections[c.id] = &c
	connectionLock.Unlock()

	return &c
}
func findConnection(id misc.ConnectionId) *Connection {
	connectionLock.Lock()
	defer connectionLock.Unlock()

	conn, _ := connections[id]

	return conn
}

func (c *Connection) Close() {
	//room.LeaveAllRooms(misc.ListenerId(c.id))
	store.RemoveConnectionInfo(c.state.MachineId, c.id)

	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Connection) OnMessageFromTopic(m pubsub.Message) {
	msg := m.(*message.Message)

	fmt.Println("-------- OnMessageFromTopic ----------")
	if msg.SenderId == c.id.ListenerId() {
		c.logger.Info("Ignoring my own message")
		return
	}
	c.logger.Info("Connection.OnMessageFromTopic", "msg", msg)
	if c.conn != nil && (msg.ReceiverId == misc.ListenerId(c.id) || msg.ReceiverId == "") {
		c.conn.Write([]byte(msg.String() + "\n"))

		if msg.Cmd == "Ping" {
			msg := message.NewMessage(msg.RoomId, c.id.ListenerId(), msg.SenderId, "Pong", map[string]interface{}{})

			pubsub.SendMessage(&msg)
		}
	}
}

func (c *Connection) joinRoom(roomId misc.RoomId) {
	c.conn.Write([]byte(">>> Joining " + roomId + "\n"))
	c.consumer = pubsub.NewConsumer(c.logger, roomId.Topic(), c)
	c.consumer.StartConsumer(&message.Message{})
	msg := message.NewMessage(roomId, c.id.ListenerId(), "", "Join", map[string]interface{}{})
	pubsub.SendMessage(&msg)
}

func (c *Connection) Run() {
	defer cleanupConnection(c.id)
	c.conn.Write([]byte(">>> Welcome " + c.id + "\n"))

	// We have a new connection, let's join the global lobby
	c.joinRoom(misc.GetGlobalLobbyId())

	c.conn.Write([]byte(">>> Ready\n"))

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
		pubsub.SendMessage(&msg)
	}

}
