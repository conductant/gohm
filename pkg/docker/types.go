package docker

import (
	_docker "github.com/fsouza/go-dockerclient"
)

type Docker struct {
	Endpoint string

	Cert string
	Key  string
	Ca   string

	docker *_docker.Client

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
	Network _docker.NetworkSettings

	DockerData *_docker.Container `json:"docker_data"`

	docker *_docker.Client
}

type AuthIdentity struct {
	_docker.AuthConfiguration
}

type ContainerControl struct {
	*_docker.Config

	// If false, the container starts up in daemon mode (as a service) - default
	RunOnce       bool                `json:"run_once,omitempty"`
	HostConfig    *_docker.HostConfig `json:"host_config"`
	ContainerName string              `json:"name,omitempty"`
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
