package main

import (
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/ds"
	"github.com/hoyle1974/chorus/machine"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"github.com/hoyle1974/chorus/store"
)

type GlobalServerState struct {
	logger         *slog.Logger
	MachineId      misc.MachineId
	MachineLease   *store.Lease
	ClientCmdTopic *pubsub.Consumer
	Dist           distributed.Dist
}

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:       logger,
		MachineId:    machine.NewMachineId("EUS"),
		MachineLease: store.NewLease(time.Duration(10) * time.Second),
		Dist:         distributed.NewDist(ds.GetConn()),
	}
	ss.Dist.Put(ss.MachineId.MachineKey(), "true", ss.MachineLease)

	ss.ClientCmdTopic = pubsub.NewConsumer(logger, ss.MachineId.ClientCmdTopic(), ss)
	ss.ClientCmdTopic.StartConsumer(&message.ClientCmd{})

	return ss
}

func (s GlobalServerState) OnMessageFromTopic(m pubsub.Message) {
	msg := m.(*message.ClientCmd)
	s.logger.Debug("Client Command", "msg", msg)

	if msg.Cmd == "ClientJoin" {
		connectionId := misc.ConnectionId(msg.ReceiverId)
		conn := findClientConnection(connectionId)
		roomId := misc.RoomId(msg.Data["RoomId"].(string))
		conn.joinRoom(roomId)
	}
}
