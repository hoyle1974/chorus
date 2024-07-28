package dbx

import (
	"context"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/misc"
)

func GetRooms() {
	q().GetRooms(context.Background())
}

func CreateRoom(roomId misc.RoomId, machineId misc.MachineId, name string, script string, destroyOnOrphan bool) {
	q().CreateRoom(context.Background(), db.CreateRoomParams{
		Uuid:            string(roomId),
		MachineUuid:     string(machineId),
		Name:            name,
		Script:          script,
		DestroyOnOrphan: destroyOnOrphan,
	})
}

func DeleteRoom(roomId misc.RoomId) {
	q().DeleteRoom(context.Background(), string(roomId))
}

func AddRoomMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	q().AddRoomMember(context.Background(), db.AddRoomMemberParams{
		RoomUuid:       string(roomId),
		ConnectionUuid: string(connectionId),
	})
}

func RemoveRoomMember(roomId misc.RoomId, connectionId misc.ConnectionId) {
	q().RemoveRoomMember(context.Background(), db.RemoveRoomMemberParams{
		RoomUuid:       string(roomId),
		ConnectionUuid: string(connectionId),
	})
}
