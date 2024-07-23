package machine

import "github.com/hoyle1974/chorus/misc"

func NewMachineId(tp string) misc.MachineId {
	return misc.MachineId("Machine:" + tp + ":" + misc.UUIDString())
}
