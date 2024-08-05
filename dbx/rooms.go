package dbx

import (
	"context"
	"time"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

type Room struct {
	Uuid            misc.RoomId
	MachineUuid     misc.MachineId
	Name            string
	Script          string
	DestroyOnOrphan bool
	CreatedAt       time.Time
	LastUpdated     time.Time
}

func toRoom(in db.Room) Room {
	return Room{
		Uuid:            misc.RoomId(in.Uuid),
		MachineUuid:     misc.MachineId(in.MachineUuid),
		Name:            in.Name,
		Script:          in.Script,
		DestroyOnOrphan: in.DestroyOnOrphan,
		CreatedAt:       in.CreatedAt.Time,
		LastUpdated:     in.LastUpdated.Time,
	}
}

func (r QueriesX) GetRooms() ([]Room, error) {
	rows, err := r.q.GetRooms(context.Background())
	rooms := []Room{}
	if err != nil {
		return rooms, err
	}
	for _, dbRoom := range rows {
		rooms = append(rooms, toRoom(dbRoom))
	}
	return rooms, err
}

func (r QueriesX) GetOrphanedRooms() ([]Room, error) {
	rows, err := r.q.GetOrphanedRooms(context.Background())
	rooms := []Room{}
	if err != nil {
		return rooms, err
	}
	for _, dbRoom := range rows {
		rooms = append(rooms, toRoom(dbRoom))
	}
	return rooms, err
}

func (r QueriesX) GetRoomsByMachine(machineId misc.MachineId) ([]Room, error) {
	rows, err := r.q.GetRoomsByMachine(context.Background(), string(machineId))
	rooms := []Room{}
	if err != nil {
		return rooms, err
	}
	for _, dbRoom := range rows {
		rooms = append(rooms, toRoom(dbRoom))
	}
	return rooms, err
}

func (r QueriesX) CreateRoom(roomId misc.RoomId, machineId misc.MachineId, name string, script string, destroyOnOrphan bool) error {
	return r.q.CreateRoom(context.Background(), db.CreateRoomParams{
		Uuid:            string(roomId),
		MachineUuid:     string(machineId),
		Name:            name,
		Script:          script,
		DestroyOnOrphan: destroyOnOrphan,
	})
}

func (r QueriesX) DeleteRoom(roomId misc.RoomId) {
	r.q.DeleteRoom(context.Background(), string(roomId))
}

func (r QueriesX) GetRoomMembers(roomId misc.RoomId) ([]misc.ConnectionId, error) {
	ret := []misc.ConnectionId{}

	rows, err := r.q.GetRoomMembers(context.Background(), string(roomId))
	if err != nil {
		return ret, err
	}

	for _, row := range rows {
		ret = append(ret, misc.ConnectionId(row))
	}
	return ret, nil
}

func (r QueriesX) AddRoomMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	r.q.AddRoomMember(context.Background(), db.AddRoomMemberParams{
		RoomUuid:       string(roomId),
		ConnectionUuid: string(connectionId),
	})
}

func (r QueriesX) RemoveRoomMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	r.q.RemoveRoomMember(context.Background(), db.RemoveRoomMemberParams{
		RoomUuid:       string(roomId),
		ConnectionUuid: string(connectionId),
	})
}

func (r QueriesX) SetRoomOwner(roomId misc.RoomId, oldOwner misc.MachineId, newOwner misc.MachineId) error {
	return r.q.SetRoomOwner(context.Background(), db.SetRoomOwnerParams{string(roomId), string(newOwner), string(oldOwner)})
}

func (r QueriesX) GetMembershipByConnection(connectionId misc.ConnectionId) ([]misc.ConnectionId, error) {
	ret := []misc.ConnectionId{}

	rows, err := r.q.GetMembershipByConnection(context.Background(), string(connectionId))
	if err == nil {
		for _, row := range rows {
			ret = append(ret, misc.ConnectionId(row))
		}
	}

	return ret, err
}
