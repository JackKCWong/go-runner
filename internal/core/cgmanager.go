package core

import (
	"github.com/containerd/cgroups"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type CGroup interface {
	Add(pid int) error
	Delete() error
}

type NoOpCGroup struct {
	// empty struct
}

func (c NoOpCGroup) Add(_ int) error {
	return nil
}

func (c NoOpCGroup) Delete() error {
	return nil
}

func NewCGroup(cgPath string, memLimitInBytes int64) (CGroup, error) {
	control, err := cgroups.New(cgroups.V1, cgroups.NestedPath(cgPath), &specs.LinuxResources{
		Memory: &specs.LinuxMemory{Limit: &memLimitInBytes},
	})

	if err != nil {
		return nil, err
	}

	return &cdc{
		control,
	}, nil
}

type cdc struct {
	cg cgroups.Cgroup
}

func (c *cdc) Add(pid int) error {
	return c.cg.Add(cgroups.Process{Pid: pid})
}

func (c *cdc) Delete() error {
	return c.cg.Delete()
}
