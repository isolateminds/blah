package containers

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ContainerCreator interface {
	GetContainer(ctx context.Context) Container
	Callback(ctx context.Context, err error) error
}

type createContainerPayload struct {
	container Container
	cb        CallbackFn
}

func (c createContainerPayload) GetContainer(ctx context.Context) Container {
	return c.container
}
func (c createContainerPayload) Callback(ctx context.Context, err error) error {
	if errdefs.IsConflict(err) {
		// this means we have a special error to handle because the container already exists
		return c.cb(ctx, needContainerRemoveErr(err))
	}
	return c.cb(ctx, err)
}
func NewCreateContainerPayload(container *Container, cb CallbackFn) ContainerCreator {
	return createContainerPayload{
		container: *container,
		cb:        cb,
	}
}

//Creates a new container with a object that has a ContainerCreator implementation
func createContainer(ctx context.Context, client *client.Client, wg *sync.WaitGroup, cc ContainerCreator) int {

	c := cc.GetContainer(ctx)

	body, err := client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        c.Image,
			Hostname:     c.Hostname,
			ExposedPorts: c.CreateNatExposedPortSet(),
			Env:          c.CreateENVKeyPair(),
		},
		&container.HostConfig{
			PortBindings: c.CreatePortBindings(),
			Mounts:       c.CreateMounts(),
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		c.Name,
	)

	if err != nil {
		return exit(wg, cc.Callback(ctx, err))
	}
	inspect, err := client.ContainerInspect(ctx, body.ID)
	if err != nil {
		return exit(wg, cc.Callback(ctx, err))
	}
	//when creating a container you don't have the id
	//we call inspect to get the id of the created container
	//so that a may be passed  via context
	c.ContainerID = inspect.ID

	ctx, err = contextWithContainer(ctx, &c)

	return exit(wg, cc.Callback(ctx, err))
}
