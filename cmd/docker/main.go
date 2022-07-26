package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/joho/godotenv"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/ttacon/chalk"
)

const (
	DefaultImageName     = "blah"
	DefaultContainerName = "blah-server"
	// default value for many flags
	NoInput = ""
)

// Write json response to stdout
type ErrorDetail struct {
	Message string `json:"message"`
}
type Aux struct {
	ID string `json:"ID"`
}
type DockerJSONWriter struct {
	Status string `json:"status"`
	Stream string `json:"stream"`
	Aux    Aux    `json:"aux"`
}

func (d *DockerJSONWriter) TagExists(tag string) bool {
	return strings.Trim(tag, "\n") != ""
}
func (d *DockerJSONWriter) Print(phase string, r io.ReadCloser) error {

	j := json.NewDecoder(r)
	for err := j.Decode(d); err != io.EOF; err = j.Decode(d) {
		if err != nil {
			return err
		}
		if d.TagExists(d.Status) {
			fmt.Printf("<%s> <%s> %s\n", chalk.Green.Color(phase), chalk.Yellow.Color("Status"), chalk.White.Color(d.Status))
		}
		if d.TagExists(d.Stream) {
			fmt.Printf("<%s> <%s> %s\n", chalk.Green.Color(phase), chalk.Yellow.Color("stream"), chalk.White.Color(d.Stream))
		}
		if d.TagExists(d.Aux.ID) {
			fmt.Printf("<%s> <%s> %s\n", chalk.Green.Color(phase), chalk.Yellow.Color("aux"), chalk.White.Color(d.Aux.ID))
		}

	}
	return nil
}
func FatalRed(format string, v ...any) {
	log.Fatalf(chalk.Red.Color(format), v...)
}
func CreateImage(ctxParent context.Context, client *client.Client, ctxPath *string, imageName *string, env *string) context.Context {
	ctx, cancel := context.WithCancel(ctxParent)

	go func() {
		defer cancel()

		buildCtx, err := archive.TarWithOptions(*ctxPath, &archive.TarOptions{})
		if err != nil {
			FatalRed("Error creating archive %s", err)
		}
		var port *string
		if *env != NoInput {
			envFile := fmt.Sprintf(".%s.env", *env)
			err := godotenv.Load(envFile)
			if err != nil {
				FatalRed("Error could not load %s %s\n", envFile, err)
			}
			envPort := os.Getenv("PORT")
			envAppName := os.Getenv("APP_NAME")

			if envPort == NoInput {
				FatalRed("Error -env was specified and %s was loaded but it contains no PORT variable", envFile)
			}
			if envAppName == NoInput {
				FatalRed("Error -env was specified and %s was loaded but it contains no APP_NAME variable", envFile)
			}
			envPort = fmt.Sprintf("%s/tcp", envPort)
			port = &envPort
		}
		buildResponse, err := client.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
			BuildArgs: map[string]*string{
				"ENV_NAME": env,
				"PORT":     port,
			},
			Tags:       []string{*imageName},
			Dockerfile: "Dockerfile",
		})

		if err != nil {
			FatalRed("Error building image %s\n", err)
		}
		var dw DockerJSONWriter
		dw.Print("BUILD", buildResponse.Body)
	}()
	return ctx
}

type ContainerConfig struct {
	Image        string
	Name         string
	Hostname     string
	Env          []string
	ExposedPorts nat.PortSet
	Mounts       []mount.Mount
}

func RunContainer(ctxParent context.Context, client *client.Client, config ContainerConfig) context.Context {

	ctx, cancel := context.WithCancel(ctxParent)

	go func() {
		defer cancel()
		containerCreateCreatedBody, err := client.ContainerCreate(
			ctx,
			&container.Config{
				Env:          config.Env,
				Hostname:     config.Hostname,
				ExposedPorts: config.ExposedPorts,
				Image:        config.Image,
			},
			&container.HostConfig{
				Mounts: config.Mounts,
			},
			&network.NetworkingConfig{},
			&v1.Platform{
				OS: "linux",
			},
			config.Name,
		)
		if err != nil {
			if err, ok := err.(errdefs.ErrConflict); ok {
				FatalRed("Container %s Already in use %s", config.Name, err)
			}
			FatalRed("Error creating container %s", err)
		}
		log.Printf("Container %s has been created %s", config.Name, containerCreateCreatedBody.ID)

		err = client.ContainerStart(ctx, containerCreateCreatedBody.ID, types.ContainerStartOptions{})
		if err != nil {
			FatalRed("Error starting container %s %s", config.Name, err)
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
	fmt.Println("\tmongo\tCreate a MongoDB container")
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
			Value: NoInput,
			Usage: "Path where Dockerfile resides",
		},
		{
			Name:  "name",
			Value: DefaultImageName,
			Usage: "A tag to use as an image name for containers to use",
		},
		{
			Name:  "env",
			Value: "",
			Usage: "Name of the .env file prefix to load",
		},
	})
	containerFlagSet, containerFlags := SetFlags(flag.NewFlagSet("container", flag.ExitOnError), []Flag{
		{
			Name:  "name",
			Value: NoInput,
			Usage: "A tag to use as an image name for containers to use",
		},
		{
			Name:  "image",
			Value: NoInput,
			Usage: "A tag to use as an image name for containers to use",
		},
		{
			Name:  "hostname",
			Value: DefaultContainerName,
			Usage: "Container hostname",
		},
		{
			Name:  "expose",
			Value: "",
			Usage: "Expose port number and protocol in the format 80/tcp",
		},
		{
			Name:  "rm",
			Value: NoInput,
			Usage: "Name of the container to remove",
		},
	})
	logsFlagSet, logFlags := SetFlags(flag.NewFlagSet("logs", flag.ExitOnError), []Flag{
		{
			Name:  "name",
			Value: NoInput,
			Usage: "Name of the container to print logs from",
		},
	})

	mongoFlagSet, mongoFlags := SetFlags(flag.NewFlagSet("mongo", flag.ExitOnError), []Flag{
		{
			Name:  "mount",
			Value: NoInput,
			Usage: "Path on the host to bind mount",
		},
	})

	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		FatalRed("Error initializing docker API client %s\n", err)
	}
	if len(os.Args) < 2 {
		PrintDefaults()
	}
	switch os.Args[1] {
	case "image":
		ParseFlags(imageFlagSet)
		ctx := context.Background()

		if *imageFlags["context"] == NoInput {
			fmt.Println("-context flag is required")
			imageFlagSet.PrintDefaults()
			os.Exit(1)
		}
		ctx = CreateImage(ctx, client, imageFlags["context"], imageFlags["name"], imageFlags["env"])
		<-ctx.Done()

	case "container":
		ParseFlags(containerFlagSet)
		ctx := context.Background()
		// bypass all and rm container if rm flag has value
		if *containerFlags["rm"] != NoInput {
			containerName := *containerFlags["rm"]
			err := client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
			if err != nil {
				FatalRed("Error removing container %s %s\n", containerName, err)
			}
		} else if *containerFlags["image"] == NoInput || *containerFlags["name"] == NoInput {
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
		if *logFlags["name"] == NoInput {
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
				FatalRed("Error Showing Log output %s\n", err)
			}
			io.Copy(os.Stdout, rc)
		}
	case "mongo":
		ParseFlags(mongoFlagSet)
		if *mongoFlags["mount"] == NoInput {
			fmt.Println("-mount flag is required")
			os.Exit(1)
		} else {
			mount := *mongoFlags["mount"]
			// an absolute path is needed for  -mount
			if !path.IsAbs(mount) {
				if mount, err = filepath.Abs(mount); err != nil {
					FatalRed("Error unable to determine absolute path from -mount input %s", err)
				} else {
					*mongoFlags["mount"] = mount
				}
			}
		}
		ctx := context.Background()
		rc, err := client.ImagePull(ctx, "mongo:latest", types.ImagePullOptions{})
		if err != nil {
			FatalRed("Error Pulling Mongo image %s", err)
		}
		var dw DockerJSONWriter
		dw.Print("PULL", rc)
		ctx = RunContainer(ctx, client, ContainerConfig{
			Image:    "mongo",
			Name:     "mongodb",
			Hostname: "database",
			Env: []string{
				"MONGO_INITDB_ROOT_USERNAME=root",
				"MONGO_INITDB_ROOT_PASSWORD=secure",
			},
			ExposedPorts: nat.PortSet{
				nat.Port("27017/tcp"): {},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: *mongoFlags["mount"],
					Target: "/data/db",
				},
			},
		})
		<-ctx.Done()

	default:
		PrintDefaults()
	}

}
