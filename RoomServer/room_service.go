package main

import (
	"encoding/json"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"github.com/redis/go-redis/v9"
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

func NewRoomInfoFromString(s string) RoomInfo {
	v := RoomInfo{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		panic(err)
	}
	return v
}

type RoomService struct {
	gss        GlobalServerState
	localRooms map[misc.RoomId]*Room
	rooms      distributed.Set
}

func StartLocalRoomService(gss GlobalServerState) *RoomService {
	rs := &RoomService{
		gss:        gss,
		localRooms: map[misc.RoomId]*Room{},
		rooms:      gss.Dist.BindSet("rooms"),
	}

	rs.Start()

	gss.logger.Info("Local Room Service is started.")
	return rs
}

func (rs *RoomService) Start() {
	go rs.run()
}

func (rs *RoomService) run() {
	for {
		roomIds, err := rs.rooms.SMembers()
		if err != nil {
			panic(err)
		}
		for _, roomId := range roomIds {
			info := rs.getRoomInfo(misc.RoomId(roomId))
			if !rs.isOwnerOnline(info.RoomId) {
				rs.gss.logger.Warn("owner is not online:", info)
				rs.waitForOwnership(info.RoomId)
				rs.bindRoomToThisMachine(info)
			}
		}

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

func (rs *RoomService) getRoomInfo(roomId misc.RoomId) RoomInfo {
	infoS, err := rs.gss.Dist.Get(roomId.RoomKey())
	if err != nil {
		return RoomInfo{}
	}
	return NewRoomInfoFromString(infoS)
}

func (rs *RoomService) isOwnerOnline(roomId misc.RoomId) bool {
	m, err := rs.gss.Dist.Get(roomId.OwnershipKey())
	if err == redis.Nil {
		return false
	}
	machineId := misc.MachineId(m)
	_, err = rs.gss.Dist.Get(machineId.MachineKey())
	if err == redis.Nil {
		return false
	}
	return true
}

func (rs *RoomService) waitForOwnership(roomId misc.RoomId) {
	rs.gss.logger.Debug("Waiting for owernship", roomId)
	for {
		b, _ := rs.gss.Dist.PutIfAbsent(roomId.OwnershipKey(), string(rs.gss.MachineId), rs.gss.MachineLease)
		if b {
			rs.gss.logger.Info("We are the owner", "room", roomId)
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

// We are the owner, but we need to bind a local struct to the
// instance in redis
func (rs *RoomService) bindRoomToThisMachine(info RoomInfo) *Room {
	rs.gss.logger.Debug("Binding locally", info.RoomId)

	rs.gss.Dist.Put(info.RoomId.RoomKey(), info.String())

	r := &Room{
		state:       rs.gss,
		roomService: rs,
		info:        info,
		logger:      rs.gss.logger.With("info", info),
		members:     rs.gss.Dist.BindSet(info.RoomId.RoomMembershipKey()),
	}
	_, err := rs.rooms.SAdd(string(info.RoomId))
	if err != nil {
		panic(err)
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
