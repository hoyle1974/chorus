package dbx

import (
	"context"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

func (r QueriesX) GetRooms() {
	r.q.GetRooms(context.Background())
}

func (r QueriesX) CreateRoom(roomId misc.RoomId, machineId misc.MachineId, name string, script string, destroyOnOrphan bool) {
	r.q.CreateRoom(context.Background(), db.CreateRoomParams{
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
