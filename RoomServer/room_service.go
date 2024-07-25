package main

import (
	"encoding/json"
	"time"

	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
)

// Room list store all room info in Redis
// When a server starts, it checks the room list to make sure all rooms are claimeed
// It may take ownership of them or clean them up based on room policies

type RoomInfo struct {
	RoomId          misc.RoomId
	Name            string
	AdminScript     string
	DestroyOnOrphan bool
}

func (r RoomInfo) String() string {
	b, e := json.Marshal(r)
	if e != nil {
		panic(e)
	}
	return string(b)
}

type RoomService struct {
	gss   GlobalServerState
	rooms map[misc.RoomId]*Room
}

func StartLocalRoomService(gss GlobalServerState) *RoomService {
	rs := &RoomService{gss: gss, rooms: map[misc.RoomId]*Room{}}

	rs.Start()

	return rs
}

func (rs *RoomService) Start() {
	go rs.run()
}

func (rs *RoomService) run() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (rs *RoomService) BootstrapLobby() {
	go func() {
		rs.waitForOwnership(misc.GetGlobalLobbyId())
		info := RoomInfo{RoomId: misc.GetGlobalLobbyId(), AdminScript: "matchmaker.js", DestroyOnOrphan: false}
		rs.bindRoomToThisMachine(info)
	}()
}

func (rs *RoomService) NewRoom(info RoomInfo) *Room {
	rs.waitForOwnership(info.RoomId)
	return rs.bindRoomToThisMachine(info)
}

func (rs *RoomService) waitForOwnership(roomId misc.RoomId) {
	rs.gss.Dist.PutIfAbsent(roomId.OwnershipKey(), string(rs.gss.MachineId), rs.gss.MachineLease)
	rs.gss.logger.Info("We are the owner", "room", roomId)
}

// We are the owner, but we need to bind a local struct to the
// instance in redis
func (rs *RoomService) bindRoomToThisMachine(info RoomInfo) *Room {
	rs.gss.Dist.Put(info.RoomId.RoomKey(), info.String())

	r := &Room{
		state:       rs.gss,
		roomService: rs,
		info:        info,
		logger:      rs.gss.logger.With("info", info),
		members:     rs.gss.Dist.BindSet(info.RoomId.RoomMembershipKey()),
	}
	ctx, err := createScriptEnvironmentForRoom(r, info.AdminScript)
	if err != nil {
		rs.gss.logger.Error("createScriptEnvironmentForRoom", "error", err)
		return nil
	}
	r.ctx = ctx

	r.consumer = pubsub.NewConsumer(r.logger, info.RoomId.Topic(), r)
	r.consumer.StartConsumer(&message.Message{})
	time.Sleep(time.Duration(1) * time.Second)

	// Ask anyone in the room to respond
	msg := message.NewMessage(info.RoomId, info.RoomId.ListenerId(), "", "Ping", map[string]interface{}{})
	pubsub.SendMessage(&msg)

	return r
}
