package containers

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
)

type ContainerStopper interface {
	GetContainerID(ctx context.Context) string
	Callback(ctx context.Context, err error) error
}
type containerStopperPayload struct {
	ID       *string
	callback CallbackFn
}

func (c containerStopperPayload) GetContainerID(ctx context.Context) string {
	if c.ID == nil {
		if container, ok := FromContainerContext(ctx); ok {
			return container.ContainerID
		}
	}
	return *c.ID
}
func (c containerStopperPayload) Callback(ctx context.Context, err error) error {
	return c.callback(ctx, err)
}

//Tries to get the container id from context if ID  param is nil
func NewContainerStopperPayload(ID *string, cb CallbackFn) ContainerStopper {
	return containerStopperPayload{ID: ID, callback: cb}

}

//Stops a running container
func stopContainer(ctx context.Context, client *client.Client, wg *sync.WaitGroup, c ContainerStopper) int {
	return exit(wg, client.ContainerStop(ctx, c.GetContainerID(ctx), nil))
}
