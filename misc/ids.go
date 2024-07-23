package misc

type TopicId string

type ConnectionId string

func (id ConnectionId) ListenerId() ListenerId {
	return ListenerId(id)
}

type RoomId string

func (id RoomId) ListenerId() ListenerId {
	return ListenerId(id)
}
func (id RoomId) Topic() TopicId {
	return TopicId(id)
}
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

const globalLobbyId = RoomId("GlobalLobby")

func GetGlobalLobbyId() RoomId {
	return globalLobbyId
}
