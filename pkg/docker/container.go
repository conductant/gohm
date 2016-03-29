package docker

import (
	"strings"
)

func (c *Container) GetComparable() string {
	return c.Image
}

func (c *Container) Inspect() error {
	cc, err := c.docker.InspectContainer(c.Id)
	if err != nil {
		return err
	}
	c.Name = cc.Name[1:] // there's this funny '/name' thing going on with how docker names containers
	c.ImageId = cc.Image
	c.Command = cc.Path + " " + strings.Join(cc.Args, " ")
	if cc.NetworkSettings != nil {
		c.Ip = cc.NetworkSettings.IPAddress
		c.Network = *cc.NetworkSettings
		c.Ports = get_ports(cc.NetworkSettings.PortMappingAPI())
	}
	c.DockerData = cc
	return nil
}
