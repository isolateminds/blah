package containers

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"gorm.io/gorm"
)

type ContainerMount struct {
	gorm.Model
	MountRefer uint
	Source     string     `json:"source"`
	Tagret     string     `json:"target"`
	Type       mount.Type `json:"type"`
}

type ContainerExposedPort struct {
	gorm.Model
	ExposedPortRefer uint
	Port             string `json:"port"`
}

//For persisting port bindings
type ContainerPortBinding struct {
	gorm.Model
	PortBindingRefer uint
	Port             string `json:"port"`
	HostPort         string `json:"hostPort"`
	HostIP           string `json:"hostIP"`
}
type ContainerEnv struct {
	gorm.Model
	EnvRefer uint
	Key      string `json:"key"`
	Value    string `json:"value"`
}

//Configuration struct for create container
type Container struct {
	gorm.Model
	ContainerID  string                 `json:"containerID"`
	Name         string                 `json:"name"`
	Image        string                 `json:"image"`
	Hostname     string                 `json:"hostname"`
	Mounts       []ContainerMount       `gorm:"foreignKey:MountRefer;       constraint:OnDelete:CASCADE;" json:"mounts"`
	ExposedPorts []ContainerExposedPort `gorm:"foreignKey:ExposedPortRefer; constraint:OnDelete:CASCADE;" json:"exposedPorts"`
	PortBindings []ContainerPortBinding `gorm:"foreignKey:PortBindingRefer; constraint:OnDelete:CASCADE;" json:"portBindings"`
	Env          []ContainerEnv         `gorm:"foreignKey:EnvRefer;         constraint:OnDelete:CASCADE;" json:"env"`
}

//Had to make different methods here because
//gorm not being able to accept some types the
//docker sdk uses

//Makes KEY=pair
func (c Container) CreateENVKeyPair() []string {
	keypair := make([]string, len(c.Env))
	for i := range c.Env {
		keypair[i] = fmt.Sprintf("%s=%s", c.Env[i].Key, c.Env[i].Value)
	}
	return keypair
}

func (c Container) CreateNatExposedPortSet() nat.PortSet {
	pSet := make(nat.PortSet)
	for i := range c.ExposedPorts {
		pSet[nat.Port(c.ExposedPorts[i].Port)] = struct{}{}
	}
	return pSet
}
func (c Container) CreatePortBindings() nat.PortMap {
	pMap := make(nat.PortMap)
	for i := range c.PortBindings {
		pMap[nat.Port(c.PortBindings[i].Port)] = []nat.PortBinding{
			{
				HostPort: c.PortBindings[i].HostPort,
				HostIP:   c.PortBindings[i].HostIP,
			},
		}
	}
	return pMap
}
func (c Container) CreateMounts() []mount.Mount {
	var mounts []mount.Mount
	for i := range c.Mounts {

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: c.Mounts[i].Source,
			Target: c.Mounts[i].Tagret,
		})
	}

	return mounts
}
