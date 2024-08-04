package main

import (
	"log/slog"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
)

type GlobalServerState struct {
	logger         *slog.Logger
	machineId      misc.MachineId
	clientCmdTopic *pubsub.Consumer
	q              dbx.QueriesX
}

func (s GlobalServerState) Logger() *slog.Logger      { return s.logger }
func (s GlobalServerState) MachineId() misc.MachineId { return s.machineId }
func (s GlobalServerState) MachineType() string       { return "EUS" }

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:    logger,
		machineId: machine.NewMachineId("EUS"),
		q:         dbx.Dbx().Queries(db.New(dbx.GetConn())),
	}

	ss.clientCmdTopic = pubsub.NewConsumer(logger, string(ss.machineId), ss.machineId.ClientCmdTopic(), ss)
	ss.clientCmdTopic.StartConsumer(&message.ClientCmd{})

	return ss
}

func (s GlobalServerState) OnMessageFromTopic(m pubsub.Message) {
	msg := m.(*message.ClientCmd)
	s.logger.Debug("Client Command", "msg", msg)

	if msg.Cmd == "ClientJoin" {
		connectionId := misc.ConnectionId(msg.ReceiverId)
		conn := findLocalClientConnection(connectionId)
		if conn == nil {
			s.logger.Warn("Tried to send a message to a local client that does not exist", "msg", msg)
			return
		}
		roomId := misc.RoomId(msg.Data["RoomId"].(string))
		conn.joinRoom(roomId)
	}
}
