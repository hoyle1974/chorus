package ownership

import (
	"log/slog"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/misc"
)

type OwnershipService struct {
	ctx            OwnershipContext
	ownership      distributed.Hash
	ownershipLease *distributed.HashLease
}

type OwnershipContext interface {
	Logger() *slog.Logger
	MachineId() misc.MachineId
	Dist() distributed.Dist
}

func StartLocalOwnershipService(ctx OwnershipContext) *OwnershipService {
	service := &OwnershipService{
		ctx:       ctx,
		ownership: ctx.Dist().BindHash("ownership"),
	}
	service.ownershipLease = service.ownership.NewHashLease(time.Duration(10) * time.Second)

	service.Start()

	service.ctx.Logger().Info("Local Ownership Service is started.")
	return service
}

func (o *OwnershipService) StopLocalService() {
	// TODO
	o.ownershipLease.Destroy()
}

func (o *OwnershipService) Start() {
	go o.run()
}

func (o *OwnershipService) run() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
	}
}

type ResourceId interface {
	String() string
}

func (o *OwnershipService) ReleaseOwnership(resourceId ResourceId) {
	o.ownership.HDel(resourceId.String())
}

func (o *OwnershipService) ClaimOwnership(resourceId ResourceId, wait time.Duration) bool {
	start := time.Now()
	for {
		owner := o.GetValidOwner(resourceId)
		if owner == misc.NilMachineId {
			// Try to become the owner
			if ok, _ := o.ownership.HSetNX(resourceId.String(), string(o.ctx.MachineId())); ok {
				// We claimed ownership, add our lease key
				o.ownershipLease.AddKey(resourceId.String())
				return true
			}
		}
		time.Sleep(time.Second)
		if time.Now().Sub(start).Seconds() >= wait.Seconds() {
			return false
		}
	}

	return false
}

func (o *OwnershipService) GetOwner(resourceId ResourceId) misc.MachineId {
	m, err := o.ownership.HGet(resourceId.String())
	if err != nil {
		return misc.NilMachineId
	}
	return misc.MachineId(m)
}

func (o *OwnershipService) GetValidOwner(resourceId ResourceId) misc.MachineId {
	owner := o.GetOwner(resourceId)
	if exists, _ := o.ctx.Dist().Exists(owner.MachineKey()); exists {
		return owner
	}

	return misc.NilMachineId
}
