package containers

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
	"github.com/isolateminds/blah/internal/color"
)

// Callback function used in all commands the error needs to be handled in the body of the callback function
type CallbackFn func(ctx context.Context, err error) error

type key int

var id key

// Creates a new context for the container it holds the container ID value
func contextWithContainer(ctx context.Context, c *Container) (context.Context, error) {
	if c.ContainerID == "" {
		return nil, noContextIDError(fmt.Errorf("Error Container ID does not have a value. %#v", c.ContainerID))
	}
	return context.WithValue(ctx, id, c), nil
}

// Retrieves the container from its context
func FromContainerContext(ctx context.Context) (*Container, bool) {
	container, ok := ctx.Value(id).(*Container)
	return container, ok
}

// A wrapper for the docker api client
type Controller struct{ client *client.Client }

// Starts a container command that implements a set of interfaces and
// has its own sync wait group and the wait group is incremented by one
// and Done() is called after the callback returns
func (c *Controller) Start(ctx context.Context, command any) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	switch command.(type) {
	case ContainerCreator:
		go createContainer(ctx, c.client, &wg, command.(ContainerCreator))
		break
	case ContainerStarter:
		go startContainer(ctx, c.client, &wg, command.(ContainerStarter))
		break
	case ContainerRemover:
		go removeContainer(ctx, c.client, &wg, command.(ContainerRemover))
		break
	case ContainerStopper:
		go stopContainer(ctx, c.client, &wg, command.(ContainerStopper))
		break
	case ImagePuller:
		go pullImage(ctx, c.client, &wg, command.(ImagePuller))
		break
	default:
		color.PrintFatal(fmt.Errorf("%T Is not a valid container command", command))
	}
	return &wg
}

func NewController(parent context.Context, client *client.Client) (*Controller, error) {
	ctx, cancel := context.WithCancel(parent)
	defer ensureCTXCanceled(ctx, cancel)
	if !engineOnline(ctx, client) {
		return nil, engineOfflineError(fmt.Errorf("Docker engine is offline. Start the docker daemon and restart the application."))
	}
	return &Controller{client}, nil
}
