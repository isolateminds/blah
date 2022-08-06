package containers

import (
	"context"

	"github.com/docker/docker/client"
)

func engineOnline(ctx context.Context, client *client.Client) bool {
	_, err := client.Info(ctx)
	return err == nil
}
