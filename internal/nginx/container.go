package nginx

import (
	_ "embed"
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/isolateminds/blah/internal/containers"
	"github.com/isolateminds/blah/internal/utils"
)

var (
	//go:embed nginx.conf
	nginxConf []byte

	DefaultImgTag = "nginx:latest"
)

func InitialSetup(projectName string, fn containers.CallbackFn) (containers.ContainerCreator, error) {
	nginxPath := utils.WriteFileAbs(nginxConf, "nginx.conf")
	container := containers.Container{
		Name:     utils.PrefixProjectName(projectName, "nginx"),
		Image:    DefaultImgTag,
		Hostname: fmt.Sprintf("com.%s.nginx", projectName),
		Mounts: []containers.ContainerMount{
			{
				Type:   mount.TypeBind,
				Source: nginxPath,
				Tagret: "/etc/nginx/nginx.conf",
			},
		},
		ExposedPorts: []containers.ContainerExposedPort{
			{
				Port: "80",
			},
		},
		PortBindings: []containers.ContainerPortBinding{
			{
				Port:     "80",
				HostPort: "8080",
				HostIP:   "0.0.0.0",
			},
		},
	}
	return containers.NewCreateContainerPayload(&container, fn), nil
}
