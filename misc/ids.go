package misc

type ConnectionId string
type RoomId string

func (id RoomId) RoomKey() string {
	return "rooms/" + string(id)
}
func (id RoomId) RoomMembershipKey() string {
	return "rooms/" + string(id) + "/membership"
}
func (id RoomId) OwnershipKey() string {
	return string("ownership/rooms/" + id)
}

type ListenerId string

type MachineId string

func (id MachineId) MachineKey() string {
	return string("machines/" + id)
}
