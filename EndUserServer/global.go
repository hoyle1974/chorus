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
	machineId      misc.MachineId
	machineLease   *store.Lease
	clientCmdTopic *pubsub.Consumer
	dist           distributed.Dist
}

func (s GlobalServerState) Logger() *slog.Logger      { return s.logger }
func (s GlobalServerState) MachineId() misc.MachineId { return s.machineId }
func (s GlobalServerState) Dist() distributed.Dist    { return s.dist }

func NewGlobalState(logger *slog.Logger) GlobalServerState {
	ss := GlobalServerState{
		logger:       logger,
		machineId:    machine.NewMachineId("EUS"),
		machineLease: store.NewLease(time.Duration(10) * time.Second),
		dist:         distributed.NewDist(ds.GetConn()),
	}
	ss.dist.Put(ss.MachineId().MachineKey(), "true", ss.machineLease)

	ss.clientCmdTopic = pubsub.NewConsumer(logger, ss.machineId.ClientCmdTopic(), ss)
	ss.clientCmdTopic.StartConsumer(&message.ClientCmd{})

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
