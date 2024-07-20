package message

import (
	"encoding/json"

	"github.com/hoyle1974/chorus/misc"
)

type Message struct {
	RoomId     misc.RoomId
	SenderId   misc.ListenerId
	ReceiverId misc.ListenerId
	Cmd        string
	Data       map[string]interface{}
}

func NewMessage(roomId misc.RoomId, senderId misc.ListenerId, receiverId misc.ListenerId, cmd string, data map[string]interface{}) Message {
	if roomId == "" {
		panic("roomId must exist")
	}
	if senderId == "" {
		panic("senderId must exist")
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	return Message{
		RoomId:     roomId,
		SenderId:   senderId,
		ReceiverId: receiverId,
		Cmd:        cmd,
		Data:       data,
	}
}

func NewMessageFromString(msg string) Message {
	var m Message
	json.Unmarshal([]byte(msg), &m)
	if m.Data == nil {
		m.Data = map[string]interface{}{}
	}
	return m
}

func NewErrorMessage(roomId misc.RoomId, senderId misc.ListenerId, err error) Message {
	data := map[string]interface{}{}
	data["err"] = err.Error()
	return NewMessage(roomId, senderId, "", "error", data)
}

func (m Message) String() string {
	jsonData, _ := json.Marshal(m)
	return string(jsonData)
}
