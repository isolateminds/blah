package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/archive"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultImageName     = "blah"
	DefaultContainerName = "blah-server"
)

func CreateImage(ctxParent context.Context, client *client.Client, ctxPath *string, imageName *string) context.Context {
	ctx, cancel := context.WithCancel(ctxParent)

	go func() {
		defer cancel()

		buildCtx, err := archive.TarWithOptions(*ctxPath, &archive.TarOptions{})
		if err != nil {
			log.Fatalf("Error creating archive %s", err)
		}

		buildResponse, err := client.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
			Tags:       []string{*imageName},
			Dockerfile: "Dockerfile",
		})
		defer buildResponse.Body.Close()

		if err != nil {
			log.Fatalf("Error building image %s\n", err)
		}
		io.Copy(os.Stdout, buildResponse.Body)
	}()
	return ctx
}

type ContainerConfig struct {
	Image    string
	Name     string
	Hostname string
}

func RunContainer(ctxParent context.Context, client *client.Client, config ContainerConfig) context.Context {

	ctx, cancel := context.WithCancel(ctxParent)

	go func() {
		defer cancel()
		containerCreateCreatedBody, err := client.ContainerCreate(
			ctx,
			&container.Config{
				Hostname: config.Hostname,
				Image:    config.Image,
			},
			&container.HostConfig{},
			&network.NetworkingConfig{},
			&v1.Platform{
				OS: "linux",
			},
			config.Name,
		)
		if err != nil {
			if err, ok := err.(errdefs.ErrConflict); ok {
				log.Fatalf("Container %s Already in use %s", config.Name, err)
			}
			log.Fatalf("Error creating container %s", err)
		}
		log.Printf("Container %s has been created %s", config.Name, containerCreateCreatedBody.ID)

		err = client.ContainerStart(ctx, containerCreateCreatedBody.ID, types.ContainerStartOptions{})
		if err != nil {
			log.Fatalf("Error starting container %s %s", config.Name, err)
		}
		log.Printf("Container %s has been started %s", config.Name, containerCreateCreatedBody.ID)

	}()
	return ctx
}

// Prints the default command set
func PrintDefaults() {
	fmt.Println("Commands:")
	fmt.Println("\timage\tCreate an image for the server")
	fmt.Println("\tcontainer\tManage containers for the server")
	fmt.Println("\tlogs\tPrint Logs to stdout.")
	os.Exit(0)
}

// Parses flags after given comand from os.Args
func ParseFlags(fs *flag.FlagSet) {
	if err := fs.Parse(os.Args[2:]); err != nil {
		fs.PrintDefaults()
		os.Exit(0)
	}
}

type Flag struct {
	Name  string
	Value string
	Usage string
}
type Flags map[string]*string

// sets the flags to a flag set then returns a mapping of the name to the value of each flag
func SetFlags(fs *flag.FlagSet, f []Flag) (*flag.FlagSet, Flags) {
	flags := make(Flags, len(f))
	for i := range f {
		flags[f[i].Name] = fs.String(f[i].Name, f[i].Value, f[i].Usage)
	}
	return fs, flags
}

func main() {
	imageFlagSet, imageFlags := SetFlags(flag.NewFlagSet("image", flag.ExitOnError), []Flag{
		{
			Name:  "context",
			Value: "",
			Usage: "Path where Dockerfile resides",
		},
		{
			Name:  "name",
			Value: DefaultImageName,
			Usage: "A tag to use as an image name for containers to use",
		},
	})
	containerFlagSet, containerFlags := SetFlags(flag.NewFlagSet("container", flag.ExitOnError), []Flag{
		{
			Name:  "name",
			Value: "",
			Usage: "A tag to use as an image name for containers to use",
		},
		{
			Name:  "image",
			Value: "",
			Usage: "A tag to use as an image name for containers to use",
		},
		{
			Name:  "hostname",
			Value: DefaultContainerName,
			Usage: "Container hostname",
		},
		{
			Name:  "rm",
			Value: "",
			Usage: "Name of the container to remove",
		},
	})
	logsFlagSet, logFlags := SetFlags(flag.NewFlagSet("logs", flag.ExitOnError), []Flag{
		{
			Name:  "name",
			Value: "",
			Usage: "Name of the container to print logs from",
		},
	})

	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Error initializing docker API client %s\n", err)
	}
	if len(os.Args) < 2 {
		PrintDefaults()
	}
	switch os.Args[1] {
	case "image":
		ParseFlags(imageFlagSet)
		ctx := context.Background()
		if *imageFlags["context"] == "" {
			fmt.Println("-context flag is required")
			imageFlagSet.PrintDefaults()
			os.Exit(1)
		}
		ctx = CreateImage(ctx, client, imageFlags["context"], imageFlags["name"])
		<-ctx.Done()

	case "container":
		ParseFlags(containerFlagSet)
		ctx := context.Background()
		// bypass all and rm container if rm flag has value
		if *containerFlags["rm"] != "" {
			containerName := *containerFlags["rm"]
			err := client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
			if err != nil {
				log.Fatalf("Error removing container %s %s\n", containerName, err)
			}
		} else if *containerFlags["image"] == "" || *containerFlags["name"] == "" {
			fmt.Println("-image flag and -name flag is required")
			containerFlagSet.PrintDefaults()
			os.Exit(1)
		} else {
			ctx = RunContainer(ctx, client, ContainerConfig{
				Name:     *containerFlags["name"],
				Hostname: *containerFlags["hostname"],
				Image:    *containerFlags["image"],
			})
			<-ctx.Done()
		}
	case "logs":
		ParseFlags(logsFlagSet)
		ctx := context.Background()
		if *logFlags["name"] == "" {
			fmt.Println("-name flag is required")
			logsFlagSet.PrintDefaults()
			os.Exit(1)
		} else {
			rc, err := client.ContainerLogs(ctx, *logFlags["name"], types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Timestamps: true,
				Follow:     true,
				Details:    true,
			})
			if err != nil {
				log.Fatalf("Error Showing Log output %s\n", err)
			}
			io.Copy(os.Stdout, rc)
		}
	default:
		PrintDefaults()
	}

}
