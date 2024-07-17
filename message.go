package main

import (
	"encoding/json"

	"github.com/lithammer/shortuuid/v4"
)

type RoomId string
type ListenerId string
type MessageId string

type Message struct {
	MsgId      MessageId
	RoomId     RoomId
	SenderId   ListenerId
	ReceiverId ListenerId
	Cmd        string
	Data       map[string]string
}

func NewMessage(roomId RoomId, senderId ListenerId, receiverId ListenerId, cmd string, data map[string]string) Message {
	if roomId == "" {
		panic("roomId must exist")
	}
	if senderId == "" {
		panic("senderId must exist")
	}
	if data == nil {
		data = map[string]string{}
	}

	return Message{
		MsgId:      MessageId(shortuuid.New()),
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
		m.Data = map[string]string{}
	}
	return m
}

func NewErrorMessage(roomId RoomId, senderId ListenerId, err error) Message {
	data := map[string]string{}
	data["err"] = err.Error()
	return NewMessage(roomId, senderId, "", "error", data)
}

func (m Message) String() string {
	jsonData, _ := json.Marshal(m)
	return string(jsonData)
}
