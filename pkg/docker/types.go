package docker

import (
	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	Endpoint string

	Cert string
	Key  string
	Ca   string

	docker *docker.Client

	ContainerCreated func(*Container)
	ContainerStarted func(*Container)
}

type Port struct {
	ContainerPort int64  `json:"container_port"`
	HostPort      int64  `json:"host_port"`
	Type          string `json:"protocol"`
	AcceptIP      string `json:"accepts_ip"`
}

type Container struct {
	Id      string `json:"id"`
	Ip      string `json:"ip"`
	Image   string `json:"image"`
	ImageId string `json:"image_id"`

	Name    string `json:"name"`
	Command string `json:"command"`
	Ports   []Port `json:"ports"`
	Network docker.NetworkSettings

	DockerData *docker.Container `json:"docker_data"`

	docker *docker.Client
}

type AuthIdentity struct {
	docker.AuthConfiguration
}

type ContainerControl struct {
	*docker.Config

	// If false, the container starts up in daemon mode (as a service) - default
	RunOnce       bool               `json:"run_once,omitempty"`
	HostConfig    *docker.HostConfig `json:"host_config"`
	ContainerName string             `json:"name,omitempty"`
}

type Action int

const (
	Create Action = iota
	Start
	Stop
	Remove
	Die
)

// Docker event status are create -> start -> die -> stop for a container then destroy for docker -rm
var verbs map[string]Action = map[string]Action{
	"create":  Create,
	"start":   Start,
	"stop":    Stop,
	"destroy": Remove,
	"die":     Die,
}
