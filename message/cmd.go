package message

import (
	"encoding/json"

	"github.com/hoyle1974/chorus/misc"
)

type ClientCmd struct {
	MachineId  misc.MachineId
	ReceiverId misc.ListenerId
	Cmd        string
	Data       map[string]interface{}
}

func NewClientCmd(machineId misc.MachineId, receiverId misc.ListenerId, cmd string, data map[string]interface{}) ClientCmd {
	if machineId == "" {
		panic("machineId must exist")
	}
	if cmd == "" {
		panic("cmd must have a value")
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	return ClientCmd{
		MachineId:  machineId,
		ReceiverId: receiverId,
		Cmd:        cmd,
		Data:       data,
	}
}

func NewClientCmdFromString(msg string) ClientCmd {
	var m ClientCmd
	json.Unmarshal([]byte(msg), &m)
	if m.Data == nil {
		m.Data = map[string]interface{}{}
	}
	return m
}

func (m *ClientCmd) String() string {
	jsonData, _ := json.Marshal(m)
	return string(jsonData)
}
func (m *ClientCmd) Topic() misc.TopicId {
	return m.MachineId.ClientCmdTopic()
}
func (m *ClientCmd) Unmarshal(payload []byte) {
	json.Unmarshal(payload, &m)
}
