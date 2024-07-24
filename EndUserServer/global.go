package main

import (
	"log/slog"

	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
)

type GlobalServerState struct {
	logger         *slog.Logger
	MachineId      misc.MachineId
	ClientCmdTopic *pubsub.Consumer
}

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:    logger,
		MachineId: machine.NewMachineId("EUS"),
	}
	ss.ClientCmdTopic = pubsub.NewConsumer(logger, ss.MachineId.ClientCmdTopic(), ss)
	ss.ClientCmdTopic.StartConsumer(&message.ClientCmd{})

	return ss
}

func (s GlobalServerState) OnMessageFromTopic(m pubsub.Message) {
	msg := m.(*message.ClientCmd)
	s.logger.Debug("Client Command", "msg", msg)

	if msg.Cmd == "ClientJoin" {
		connectionId := misc.ConnectionId(msg.ReceiverId)
		conn := findConnection(connectionId)
		roomId := misc.RoomId(msg.Data["RoomId"].(string))
		conn.joinRoom(roomId)
	}
}
