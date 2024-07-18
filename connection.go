package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
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
	if c.conn == nil {
		return
	}
	c.conn.Write([]byte(msg.String() + "\n"))
}

func NewConnection(conn net.Conn) *Connection {
	c := Connection{
		id:   ListenerId("L" + UUIDString()),
		conn: conn,
	}

	connectionLock.Lock()
	defer connectionLock.Unlock()
	connections[c.id] = &c

	return &c
}

func (c *Connection) Close() {
	fmt.Println("Closing connection for ", c.id)

	// Leave all rooms
	roomLock.Lock()
	for _, room := range rooms {
		if room.HasListener(c.id) {
			room.leave(c.id)
		}
	}
	roomLock.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	connectionLock.Lock()
	delete(connections, c.id)
	connectionLock.Unlock()

}

func (c *Connection) Run(room *Room) {
	defer c.Close()

	// We have a new connection, let's join the global lobby
	room.Join(c.id, c)

	buf := make([]byte, 65536)
	for {
		n, err := c.conn.Read(buf)
		if err == io.EOF {
			fmt.Println("Disconnect user ", c.id)
			c.conn = nil
			return
		}
		if err != nil {
			fmt.Println("*** ", err)
			c.conn.Write([]byte(NewErrorMessage(room.id, c.id, err).String() + "\n"))
			return
		}
		// Trim trailing newline (if present)
		words := strings.Fields(strings.TrimSpace(string(buf[:n])))
		if len(words) == 1 {
			if words[0] == "exit" {
				return
			}

			if words[0] == "info" {
				info := map[string]interface{}{}

				roomList := []interface{}{}
				roomLock.Lock()
				for roomId, room := range rooms {
					if room.HasListener(c.id) {
						roomList = append(roomList,
							map[string]interface{}{
								"roomId":    roomId,
								"listeners": len(room.listeners),
								"script":    room.script,
								"name":      room.name,
							},
						)
					}
				}
				roomLock.Unlock()

				listenerList := []interface{}{}
				connectionLock.Lock()
				for listenerId, _ := range connections {
					listenerList = append(listenerList,
						map[string]interface{}{
							"id": listenerId,
						},
					)
				}
				connectionLock.Unlock()

				info["rooms"] = roomList
				info["connections"] = listenerList

				c.OnMessage(NewMessage("none", "none", c.id, "info", info))
				continue
			}
		}

		if len(words) < 2 {
			continue
		}
		msg := NewMessage(RoomId(words[0]), c.id, "room", string(words[1]), map[string]interface{}{})

		for t := 0; t < len(words)-2; t += 2 {
			key := words[t+2]
			value := words[t+3]
			msg.Data[key] = value
		}

		room := FindRoom(msg.RoomId)
		if room != nil {
			room.Send(msg)
		}
	}

}
