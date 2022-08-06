package mongodb

import (
	"bufio"
	_ "embed"
	"net/url"
	"os"

	"context"
	"fmt"

	"github.com/dannyvidal/blah/internal/color"
	"github.com/dannyvidal/blah/internal/containers"
	"github.com/dannyvidal/blah/internal/utils"
	"github.com/docker/docker/api/types/mount"
)

var (
	//go:embed init-mongo.sh
	initdbFile           []byte
	defaultHostPort      = "3186"
	defaultEntrypoint    = "initdb"
	defaultContainerName = "mongo"

	DefaultImgTag = "mongo:latest"
)

//Container creation object spefically for mongodb image

type startContainerPayload struct {
	ID string
}

func (p startContainerPayload) HandleAfterStart(ctx context.Context) error {
	fmt.Printf("Container %s has been started\n", p.ID)
	return nil
}
func (p startContainerPayload) GetStartOptions() containers.ContainerStartOptions {
	return containers.ContainerStartOptions{ID: p.ID}
}

//Prompts user for mongodb authentication details and makes all necessary mount points
func InitialSetup(projectName string, cb containers.CallbackFn) (containers.ContainerCreator, error) {

	var (
		user  string
		pass  string
		passC string
	)
	color.PrintStatus("MongoDB", "Setup your Mongo database")
	reader := bufio.NewReader(os.Stdin)

	err := utils.UntilError(
		func() error {
			return utils.GetInput(reader, "Username: ", &user, false, "username")
		},
		func() error {
			return utils.GetInput(reader, "Password: ", &pass, true, "password")
		},
		func() error {
			return utils.GetInput(reader, "Confirm Password: ", &passC, true, "password confirmation")
		},
	)

	if err != nil {
		return nil, err
	}

	if pass != passC {
		return nil, fmt.Errorf("Passwords do not match")
	}
	//Create Database mountpoints
	initdbDir := utils.MkdirAbs(defaultEntrypoint)
	initdbDirPath := utils.GetAbsChild(initdbDir)
	utils.WriteFile(initdbFile, initdbDir, "init-db.sh")
	databasePath := utils.MkdirAbs("database")

	//Root password 16 char long string, saves user time, to not think about two separate passwords
	rootPass := utils.GenerateRandomString(16)
	URL := fmt.Sprintf("mongodb://%s:%s@localhost:%s/%s", user, url.QueryEscape(pass), defaultHostPort, projectName)
	rootURL := fmt.Sprintf("mongodb://%s:%s@localhost:%s/%s", "root", rootPass, defaultHostPort, "admin")

	container := containers.Container{
		Name:     utils.PrefixProjectName(projectName, "mongodb"),
		Hostname: fmt.Sprintf("com.%s.mongodb", projectName),
		Image:    DefaultImgTag,
		Env: []containers.ContainerEnv{
			{
				Key:   "MONGO_INITDB_DATABASE",
				Value: projectName,
			},
			{
				Key:   "MONGO_INITDB_USERNAME",
				Value: user,
			},
			{
				Key:   "MONGO_INITDB_PASSWORD",
				Value: pass,
			},
			{
				Key:   "MONGODB_URL",
				Value: URL,
			},
			{
				Key:   "MONGO_INITDB_ROOT_USERNAME",
				Value: "root",
			},
			{
				Key:   "MONGO_INITDB_ROOT_PASSWORD",
				Value: rootPass,
			},

			{
				Key:   "MONGODB_ROOT_URL",
				Value: rootURL,
			},
		},
		Mounts: []containers.ContainerMount{
			{
				Type:   mount.TypeBind,
				Source: initdbDirPath,
				Tagret: "/docker-entrypoint-initdb.d",
			},
			{
				Type:   mount.TypeBind,
				Source: databasePath,
				Tagret: "/data/db",
			},
		},
		PortBindings: []containers.ContainerPortBinding{
			{
				Port:     "27017",
				HostPort: defaultHostPort,
				HostIP:   "0.0.0.0",
			},
		},
		ExposedPorts: []containers.ContainerExposedPort{
			{
				Port: "27017",
			},
		},
	}
	//Eg. MONGO_INITDB_USERNAME=admin
	utils.AppendFileIfNotExists(".env", container.CreateENVKeyPair()...)

	return containers.NewCreateContainerPayload(&container, cb), nil
}
