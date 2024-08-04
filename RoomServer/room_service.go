package main

import (
	"encoding/json"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
)

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
	state      GlobalServerState
	localRooms map[misc.RoomId]*Room
}

func (rs *RoomService) AddMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	q.AddRoomMember(roomId, connectionId)
}

func (rs *RoomService) RemoveMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	q.RemoveRoomMember(roomId, connectionId)
}

func (rs *RoomService) DeleteRoom(roomId misc.RoomId) {
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	members, err := q.GetRoomMembers(roomId)
	if err != nil {
		for _, member := range members {
			rs.RemoveMember(roomId, member)
		}
	}
	pubsub.DeleteTopic(roomId.Topic())
}

func StartLocalRoomService(state GlobalServerState) *RoomService {
	rs := &RoomService{
		state:      state,
		localRooms: map[misc.RoomId]*Room{},
	}

	state.logger.Info("Local Room Service is started.")
	return rs
}

func (rs *RoomService) RoomServiceProcess() {
	for {
		q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
		rooms, err := q.GetOrphanedRooms()
		if err != nil {
			rs.state.logger.Error("Could not get rooms", "error", err)
			return
		}
		for _, dbRoom := range rooms {
			rs.state.logger.Info("Orphaned Room", "room", dbRoom)
			/*
				for _, dbRoom := range rooms {
					rs.state.logger.Warn("owner is not online:", "info", dbRoom)
					if rs.state.ownership.ClaimOwnership(dbRoom.RoomId, time.Duration(15)*time.Second) {
						room := rs.bindRoomToThisMachine(dbRoom)
						if dbRoom.DestroyOnOrphan {
							room.Destroy()
						}
					} else {
						rs.state.logger.Warn("could not claim ownership", "info", dbRoom)
					}
				}
			*/
		}
	}
}

func (rs *RoomService) BootstrapLobby() bool {
	info := RoomInfo{
		RoomId:          misc.GetGlobalLobbyId(),
		AdminScript:     "matchmaker.js",
		Name:            "Global Lobby",
		DestroyOnOrphan: false,
	}
	_, err := rs.NewRoom(info)
	return err == nil
}

func (rs *RoomService) NewRoom(info RoomInfo) (*Room, error) {
	rs.state.logger.Debug("NewRoom", "info", info)
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	err := q.CreateRoom(info.RoomId, rs.state.MachineId(), info.Name, info.AdminScript, info.DestroyOnOrphan)
	if err != nil {
		return nil, err
	}
	pubsub.CreateTopic(info.RoomId.Topic())
	return rs.bindRoomToThisMachine(info), nil
}

// We are the owner, but we need to bind a local struct to the
// instance in redis
func (rs *RoomService) bindRoomToThisMachine(info RoomInfo) *Room {
	rs.state.logger.Debug("Binding locally", "roomId", info.RoomId)

	r := &Room{
		state:       rs.state,
		roomService: rs,
		info:        info,
		logger:      rs.state.logger.With("info", info),
	}

	ctx, err := createScriptEnvironmentForRoom(r, info.AdminScript)
	if err != nil {
		rs.state.logger.Error("createScriptEnvironmentForRoom", "error", err)
		return nil
	}
	r.ctx = ctx

	r.consumer = pubsub.NewConsumer(r.logger, string(rs.state.machineId), info.RoomId.Topic(), r)
	r.consumer.StartConsumer(&message.Message{})
	time.Sleep(time.Duration(1) * time.Second)

	// Ask anyone in the room to respond
	msg := message.NewMessage(info.RoomId, info.RoomId.ListenerId(), "", "Ping", map[string]interface{}{})
	pubsub.SendMessage(&msg)

	return r
}
