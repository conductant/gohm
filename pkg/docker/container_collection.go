package docker

import (
	"fmt"
	"github.com/conductant/gohm/pkg/fsm"
)

// For all instances of a same container image version.
type ContainerCollection struct {
	Image   string
	FsmById map[string]*fsm.Fsm
}

func (c ContainerCollection) GetKey() interface{} {
	return c.Image
}

func (c ContainerCollection) String() string {
	return fmt.Sprintf("%s (%d) instances", c.Image, len(c.FsmById))
}

func (c ContainerCollection) Instances() []*fsm.Fsm {
	list := make([]*fsm.Fsm, 0)
	for _, v := range c.FsmById {
		list = append(list, v)
	}
	return list
}

func NewContainerCollection(image string) *ContainerCollection {
	return &ContainerCollection{
		Image:   image,
		FsmById: make(map[string]*fsm.Fsm),
	}
}

func (cg *ContainerCollection) Add(c *Container) *fsm.Fsm {
	if fsm, has := cg.FsmById[c.Id]; has {
		return fsm
	} else {
		newFsm := ContainerFsm.Instance(Created)
		newFsm.CustomData = c.Id
		cg.FsmById[c.Id] = newFsm
		return newFsm
	}
}

func (cg *ContainerCollection) Remove(c *Container) {
	if _, has := cg.FsmById[c.Id]; has {
		delete(cg.FsmById, c.Id)
	}
}

func (cg *ContainerCollection) Empty() bool {
	return len(cg.FsmById) == 0
}
