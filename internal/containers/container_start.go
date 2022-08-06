package containers

import (
	"context"
	"regexp"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const portReallocationRGX = `Bind.*(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}\W.*`

type ContainerStartOptions struct {
	ID            string
	CheckpointID  string
	CheckpointDir string
}
type ContainerStarter interface {
	GetStartOptions() ContainerStartOptions
	Callback(ctx context.Context, err error) error
}
type startContainerPayload struct {
	options ContainerStartOptions
	cb      CallbackFn
}

func (p startContainerPayload) GetStartOptions() ContainerStartOptions {
	return p.options
}
func (p startContainerPayload) Callback(ctx context.Context, err error) error {
	return p.cb(ctx, err)
}

func NewStartContainerPayload(ID string, cb CallbackFn) ContainerStarter {
	if cb == nil {
		return startContainerPayload{
			options: ContainerStartOptions{ID: ID},
			cb:      func(ctx context.Context, err error) error { return err },
		}
	}
	return startContainerPayload{options: ContainerStartOptions{ID: ID}, cb: cb}
}

//Starts a container with a object that has a ContainerStarter implementation.
func startContainer(ctx context.Context, client *client.Client, wg *sync.WaitGroup, c ContainerStarter) int {
	opt := c.GetStartOptions()
	err := client.ContainerStart(ctx, opt.ID, types.ContainerStartOptions{
		CheckpointID:  opt.CheckpointID,
		CheckpointDir: opt.CheckpointDir,
	})

	if err != nil {
		if errdefs.IsSystem(err) {
			rgx := regexp.MustCompile(portReallocationRGX)
			errBytes := []byte(err.Error())
			if rgx.Match(errBytes) {
				return exit(wg, c.Callback(ctx, needPortReallocation(err)))
			}
		}
		if errdefs.IsNotFound(err) {
			//This can mean the once created container possibly needs to be re created
			return exit(wg, c.Callback(ctx, needContainerReCreate(err)))
		}
	}
	ctx, err = contextWithContainer(ctx, &Container{ContainerID: opt.ID})
	return exit(wg, c.Callback(ctx, err))
}
