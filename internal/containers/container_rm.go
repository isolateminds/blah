package containers

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

//Options for removing a container
type CRMOptions struct {
	RemoveVolumes bool
	RemoveLinks   bool
	Force         bool
	name          string
}

//A function that is called at the end of removing a container
type ContainerRemover interface {
	GetRMOptions() CRMOptions
	Callback(ctx context.Context, err error) error
}
type removeContainerPayload struct {
	options CRMOptions
	cb      CallbackFn
}

func (c removeContainerPayload) GetRMOptions() CRMOptions { return c.options }
func (c removeContainerPayload) Callback(ctx context.Context, err error) error {
	if errdefs.IsNotFound(err) {
		//TODO ??
		return c.cb(ctx, err)
	}
	return c.cb(ctx, err)
}

// Passes the name and callback function value into a payload object for use with the controllers API Eg.
func NewRemoveContainerPayload(name string, opt CRMOptions, cb CallbackFn) ContainerRemover {
	return removeContainerPayload{
		options: CRMOptions{opt.RemoveVolumes, opt.RemoveLinks, opt.Force, name},
		cb:      cb,
	}
}

//Removes a new container with a object that has a ContainerRemover implementation.
func removeContainer(ctx context.Context, client *client.Client, wg *sync.WaitGroup, c ContainerRemover) int {
	options := c.GetRMOptions()
	err := client.ContainerRemove(ctx, options.name, types.ContainerRemoveOptions{
		RemoveVolumes: options.RemoveVolumes,
		RemoveLinks:   options.RemoveLinks,
		Force:         options.Force,
	})

	return exit(wg, c.Callback(ctx, err))
}
