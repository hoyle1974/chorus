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
	q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
	err := q.CreateRoom(info.RoomId, rs.state.MachineId(), info.Name, info.AdminScript, info.DestroyOnOrphan)
	if err != nil {
		return nil, err
	}
	return rs.bindRoomToThisMachine(info), nil
}

// func (rs *RoomService) getRoomInfo(roomId misc.RoomId) RoomInfo {
// 	infoS, err := rs.state.dist.Get(roomId.RoomKey())
// 	if err != nil {
// 		return RoomInfo{}
// 	}
// 	return NewRoomInfoFromString(infoS)
// }

/*
func (rs *RoomService) isOwnerOnline(roomId misc.RoomId) bool {
	m, err := rs.gss.dist.Get(roomId.OwnershipKey())
	if err == redis.Nil {
		return false
	}
	machineId := misc.MachineId(m)
	_, err = rs.gss.dist.Get(machineId.MachineKey())
	if err == redis.Nil {
		return false
	}
	return true
}
*/

/*
func (rs *RoomService) waitForOwnership(roomId misc.RoomId) {
	rs.gss.logger.Debug("Waiting for owernship", "roomId", roomId)
	for {
		b, _ := rs.gss.dist.PutIfAbsent(roomId.OwnershipKey(), string(rs.gss.machineId), rs.gss.machineLease)
		if b {
			rs.gss.logger.Info("We are the owner", "roomId", roomId)
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}
*/

// We are the owner, but we need to bind a local struct to the
// instance in redis
func (rs *RoomService) bindRoomToThisMachine(info RoomInfo) *Room {
	rs.state.logger.Debug("Binding locally", "roomId", info.RoomId)

	r := &Room{
		state:       rs.state,
		roomService: rs,
		info:        info,
		logger:      rs.state.logger.With("info", info),
		// members:     rs.state.dist.BindSet(info.RoomId.RoomMembershipKey()),
	}
	// _, err := rs.rooms.SAdd(string(info.RoomId))
	// if err != nil {
	// 	panic(err)
	// }

	ctx, err := createScriptEnvironmentForRoom(r, info.AdminScript)
	if err != nil {
		rs.state.logger.Error("createScriptEnvironmentForRoom", "error", err)
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
