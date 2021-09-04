package cg

import "github.com/containerd/cgroups"

type CgManager interface {
	Add(pid int) error
}

func NewCgManager(cgPath string) (CgManager, error) {
	control, err := cgroups.Load(cgroups.V1, cgroups.StaticPath(cgPath))
	if err != nil {
		return nil, err
	}

	return cdc{
		control,
	}, nil
}

type cdc struct {
	cg cgroups.Cgroup
}

func (c cdc) Add(pid int) error {
	return c.cg.Add(cgroups.Process{Pid: pid})
}
