package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/dannyvidal/blah/internal/color"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Image struct {
	Name string `json:"name"`
}

var DefaultImagePullWriter = &ImagePullResponse{}

type ImagePullResponse struct {
	Id             string         `json:"id"`
	Status         string         `json:"status"`
	Progress       string         `json:"progress"`
	ProgressDetail ProgressDetail `json:"progressDetail"`
	io.Writer
}
type ProgressDetail struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

func (ip *ImagePullResponse) Write(p []byte) (n int, err error) {
	jd := json.NewDecoder(bytes.NewReader(p))
	defer func(status string, progress string) {
		if status != ip.Status && status != "" {
			color.PrintStatusWithNewLine(status, progress)
		}
	}(ip.Status, ip.Progress)

	return len(p), jd.Decode(ip)
}

type ImagePuller interface {
	GetImage() Image
	Callback(ctx context.Context, err error) error
	io.Writer
}

type imagePullPayload struct {
	image    *Image
	callback CallbackFn
	writer   io.Writer
}

func (imp imagePullPayload) GetImage() Image {
	return *imp.image
}
func (imp imagePullPayload) Callback(ctx context.Context, err error) error {
	return imp.callback(ctx, err)
}
func (imp imagePullPayload) Write(b []byte) (n int, err error) {
	return imp.writer.Write(b)
}

func NewImagePullPayload(image *Image, writer io.Writer, cb CallbackFn) ImagePuller {
	if cb == nil {
		return imagePullPayload{
			image:    image,
			callback: func(ctx context.Context, err error) error { return err },
			writer:   writer,
		}
	}
	return imagePullPayload{
		image:    image,
		callback: cb,
		writer:   writer,
	}
}

func pullImage(ctx context.Context, client *client.Client, wg *sync.WaitGroup, p ImagePuller) int {
	image := p.GetImage()
	rc, err := client.ImagePull(ctx, image.Name, types.ImagePullOptions{})
	if err != nil {
		return exit(wg, err)
	}
	_, err = io.Copy(p, rc)
	return exit(wg, err)
}
