package mysql

import (
	"bufio"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/isolateminds/blah/internal/color"
	"github.com/isolateminds/blah/internal/containers"
	"github.com/isolateminds/blah/internal/utils"
)

var (
	defaultHostPort      = "3186"
	defaultContainerName = "mysql"

	DefaultImgTag = "mysql:latest"
)

// Prompts user for mysql authentication details and makes all necessary mount points
func InitialSetup(projectName string, cb containers.CallbackFn) (containers.ContainerCreator, error) {

	var (
		user  string
		pass  string
		passC string
	)
	color.PrintStatus("Mysql", "Setup your Mysql database")
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
	databasePath := utils.MkdirAbs("database")

	//Root password 16 char long string, saves user time, to not think about two separate passwords
	rootPass := utils.GenerateRandomString(16)

	container := containers.Container{
		Name:     utils.PrefixProjectName(projectName, "mysql"),
		Hostname: fmt.Sprintf("com.%s.mysql", projectName),
		Image:    DefaultImgTag,
		Env: []containers.ContainerEnv{
			{
				Key:   "MYSQL_DATABASE",
				Value: projectName,
			},
			{
				Key:   "MYSQL_USER",
				Value: user,
			},
			{
				Key:   "MYSQL_PASSWORD",
				Value: pass,
			},

			{
				Key:   "MYSQL_ROOT_PASSWORD",
				Value: rootPass,
			},
			{
				Key:   "MYSQL_PORT",
				Value: defaultHostPort,
			},
		},
		Mounts: []containers.ContainerMount{
			{
				Type:   mount.TypeBind,
				Source: databasePath,
				Tagret: "/var/lib/mysql",
			},
		},
		PortBindings: []containers.ContainerPortBinding{
			{
				Port:     "3306",
				HostPort: defaultHostPort,
				HostIP:   "0.0.0.0",
			},
		},
		ExposedPorts: []containers.ContainerExposedPort{
			{
				Port: "3306",
			},
		},
	}
	//Eg. MYSQL_USER=admin
	utils.AppendFileIfNotExists(".env", container.CreateENVKeyPair()...)

	return containers.NewCreateContainerPayload(&container, cb), nil
}
