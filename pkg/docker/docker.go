package docker

import (
	"bufio"
	"bytes"
	_docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"
	"io"
	"time"
)

// Endpoint and file paths
func NewTLSClient(endpoint string, cert, key, ca string) (c *Docker, err error) {
	c = &Docker{Endpoint: endpoint, Cert: cert, Ca: ca, Key: key}
	c.docker, err = _docker.NewTLSClient(endpoint, cert, key, ca)
	return c, err
}

func NewClient(endpoint string) (c *Docker, err error) {
	c = &Docker{Endpoint: endpoint}
	c.docker, err = _docker.NewClient(endpoint)
	return c, err
}

func (c *Docker) ListContainers() ([]*Container, error) {
	return c.FindContainers(nil)
}

func (c *Docker) FindContainersByName(name string) ([]*Container, error) {
	found := make([]*Container, 0)
	l, err := c.FindContainers(map[string][]string{
		"name": []string{name},
	})
	if err != nil {
		return nil, err
	}
	for _, cc := range l {
		err := cc.Inspect() // populates the Name, etc.
		glog.V(100).Infoln("Inspect container", *cc, "Err=", err)
		if err == nil && cc.Name == name {
			found = append(found, cc)
		}
	}
	return found, nil
}

func (c *Docker) FindContainers(filter map[string][]string) ([]*Container, error) {
	options := _docker.ListContainersOptions{
		All:  true,
		Size: true,
	}
	if filter != nil {
		options.Filters = filter
	}
	l, err := c.docker.ListContainers(options)
	if err != nil {
		return nil, err
	}
	out := []*Container{}
	for _, cc := range l {

		glog.V(100).Infoln("Matching", options, "Container==>", cc.Ports)
		c := &Container{
			Id:      cc.ID,
			Image:   cc.Image,
			Command: cc.Command,
			Ports:   get_ports(cc.Ports),
			docker:  c.docker,
		}
		c.Inspect()
		out = append(out, c)
	}
	return out, nil
}

func (c *Docker) PullImage(auth *AuthIdentity, image *Image) (<-chan error, error) {
	output_buff := bytes.NewBuffer(make([]byte, 1024*4))
	output := bufio.NewWriter(output_buff)

	err := c.docker.PullImage(_docker.PullImageOptions{
		Repository:   image.Repository,
		Registry:     image.Registry,
		Tag:          image.Tag,
		OutputStream: output,
	}, auth.AuthConfiguration)

	if err != nil {
		return nil, err
	}

	// Since the api doesn't have a channel, all we can do is read from the input
	// and then send a done signal when the input stream is exhausted.
	stopped := make(chan error)
	go func() {
		for {
			_, e := output_buff.ReadByte()
			if e == io.EOF {
				stopped <- nil
				return
			} else {
				stopped <- e
				return
			}
		}
	}()
	return stopped, err
}

func (c *Docker) StartContainer(auth *AuthIdentity, ct *ContainerControl) (*Container, error) {
	opts := _docker.CreateContainerOptions{
		Name:       ct.ContainerName,
		Config:     ct.Config,
		HostConfig: ct.HostConfig,
	}

	daemon := !ct.RunOnce
	// Detach mode (-d option in docker run)
	if daemon {
		opts.Config.AttachStdin = false
		opts.Config.AttachStdout = false
		opts.Config.AttachStderr = false
		opts.Config.StdinOnce = false
	}

	cc, err := c.docker.CreateContainer(opts)
	if err != nil {
		return nil, err
	}

	container := &Container{
		Id:     cc.ID,
		Image:  ct.Image,
		docker: c.docker,
	}

	if c.ContainerCreated != nil {
		c.ContainerCreated(container)
	}

	err = c.docker.StartContainer(cc.ID, ct.HostConfig)
	if err != nil {
		return nil, err
	}

	if c.ContainerStarted != nil {
		c.ContainerStarted(container)
	}

	err = container.Inspect()
	return container, err
}

func (c *Docker) StopContainer(auth *AuthIdentity, id string, timeout time.Duration) error {
	return c.docker.StopContainer(id, uint(timeout.Seconds()))
}

func (c *Docker) RemoveContainer(auth *AuthIdentity, id string, removeVolumes, force bool) error {
	return c.docker.RemoveContainer(_docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: removeVolumes,
		Force:         force,
	})
}

func (c *Docker) RemoveImage(image string, force, prune bool) error {
	return c.docker.RemoveImageExtended(image, _docker.RemoveImageOptions{
		Force:   force,
		NoPrune: !prune,
	})
}

func (c *Docker) WatchContainer(notify func(Action, *Container)) (chan<- bool, error) {
	return c.WatchContainerMatching(func(Action, *Container) bool { return true }, notify)
}

func (c *Docker) WatchContainerMatching(accept func(Action, *Container) bool, notify func(Action, *Container)) (chan<- bool, error) {
	stop := make(chan bool, 1)
	events := make(chan *_docker.APIEvents)
	err := c.docker.AddEventListener(events)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case event := <-events:
				glog.V(100).Infoln("Docker event:", event)

				action, has := verbs[event.Status]
				if !has {
					continue
				}

				container := &Container{Id: event.ID, Image: event.From, docker: c.docker}
				if action != Remove {
					err := container.Inspect()
					if err != nil {
						glog.Warningln("Error inspecting container", event.ID)
						continue
					}
				}

				if notify != nil && accept(action, container) {
					notify(action, container)
				}

			case done := <-stop:
				if done {
					glog.Infoln("Watch terminated.")
					return
				}
			}
		}
	}()
	return stop, nil
}

func get_ports(list []_docker.APIPort) []Port {
	out := make([]Port, len(list))
	for i, p := range list {
		out[i] = Port{
			ContainerPort: p.PrivatePort,
			HostPort:      p.PublicPort,
			Type:          p.Type,
			AcceptIP:      p.IP,
		}
	}
	return out
}
