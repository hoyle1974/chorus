package misc

var NilMachineId MachineId = MachineId("nil")
var NilListenerId ListenerId = ListenerId("nil")
var NilRoomId RoomId = RoomId("nil")

type TopicId string

type ConnectionId string

func (id ConnectionId) ListenerId() ListenerId {
	return ListenerId(id)
}

type RoomId string

func (id RoomId) String() string {
	return string(id)
}

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
func (id MachineId) ClientCmdTopic() TopicId {
	return TopicId("topics/machines/" + id + "/clientcmd")
}

const globalLobbyId = RoomId("GlobalLobby")

func GetGlobalLobbyId() RoomId {
	return globalLobbyId
}
