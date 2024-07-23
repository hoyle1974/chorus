package main

import "github.com/hoyle1974/chorus/misc"

func OwnershipKey(roomId misc.RoomId) string {
	return string("ownership/rooms/" + roomId)
}

func RoomKey(roomId misc.RoomId) string {
	return "rooms/" + string(roomId)
}
