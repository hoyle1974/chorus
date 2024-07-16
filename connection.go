package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
)

type Connection struct {
	id   string
	conn net.Conn
}

var connectionLock sync.Mutex
var connections = map[string]*Connection{}

func FindConnectionById(id string) *Connection {
	connectionLock.Lock()
	defer connectionLock.Unlock()

	return connections[id]
}

// OnMessage implements RoomListener.
func (c *Connection) OnMessage(msg Message) {
	fmt.Println("OnMessage: ", msg)
	c.conn.Write([]byte(msg.String() + "\n"))
}

type Message struct {
	MsgId      string
	SenderId   string
	ReceiverId string
	Cmd        string
	Data       map[string]string
}

func NewMessage(senderId string, receiverId string, cmd string, data map[string]string) Message {
	id, _ := uuid.NewUUID()

	return Message{
		MsgId:      id.String(),
		SenderId:   senderId,
		ReceiverId: receiverId,
		Cmd:        cmd,
		Data:       data,
	}
}

func NewMessageFromString(msg string) Message {
	var m Message
	json.Unmarshal([]byte(msg), &m)
	return m
}

func NewErrorMessage(senderId string, err error) Message {
	data := map[string]string{}
	data["err"] = err.Error()
	return NewMessage(senderId, "", "error", data)
}

func (m Message) String() string {
	jsonData, _ := json.Marshal(m)
	return string(jsonData)
}

func NewConnection(conn net.Conn) *Connection {
	id, err := uuid.NewUUID()
	fmt.Println(id)
	if err != nil {
		return nil
	}
	c := Connection{
		id:   id.String(),
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

	msg := NewMessage(c.id, "", "init", map[string]string{})
	c.conn.Write([]byte(msg.String() + "\n"))

	room.Join(c.id, c)

	buf := make([]byte, 65536)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			c.conn.Write([]byte(NewErrorMessage(c.id, err).String() + "\n"))
			return
		}

		// Trim trailing newline (if present)
		data := buf[:n]
		msg := NewMessage(c.id, "", string(data), map[string]string{})

		room.Send(msg)
	}

}
