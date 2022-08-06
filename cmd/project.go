package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/dannyvidal/blah/internal/color"
	"github.com/dannyvidal/blah/internal/containers"
	"github.com/dannyvidal/blah/internal/mongodb"
	"github.com/dannyvidal/blah/internal/mysql"
	"github.com/dannyvidal/blah/internal/nginx"
	"github.com/dannyvidal/blah/internal/persistence"
	"github.com/dannyvidal/blah/internal/utils"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const (
	MONGO_SEL = iota
	MYSQL_SEL
)

var (
	dir        string
	start      bool
	projectCmd = &cobra.Command{
		Use:     "project",
		Short:   "Manage/Create a new or existing project",
		Example: "blah project --init /path/to/project",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			if dir != "" {
				setupProject(ctx, dir)
				return
			}
			if start {
				err := godotenv.Load()
				if err != nil {
					color.PrintFatal(errors.New("Could not load .env file are you in project (root) directory?"))
				}
				startProject(ctx)
				return
			}

		},
	}
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.PersistentFlags().StringVar(&dir, "init", "", "/path/to/new/project")
	projectCmd.PersistentFlags().BoolVar(&start, "start", false, "Launch a development server.")
}

func startProject(ctx context.Context) {
	if !utils.FileExists("persist.db") {
		color.PrintFatal(fmt.Errorf("Could not find persistent database file. Are you in the project (root) directory"))
	}
	color.PrintStatus("Container", "Starting....")

	pController, err := persistence.NewPersistedDataController("persist.db")
	if err != nil {
		color.PrintFatal(err)
	}
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		color.PrintFatal(err)
	}
	cController, err := containers.NewController(ctx, client)
	conSlice, err := pController.GetAllContainers()
	for i := range conSlice {
		starter := containers.NewStartContainerPayload(conSlice[i].ContainerID, nil)
		cController.Start(ctx, starter).Wait()
		color.PrintStatus("Container", fmt.Sprintf("Started %s", conSlice[i].Name))
	}

	//Listen until SIGINT
	utils.HandleSIGTERM(func() {
		fmt.Println()
		for i := range conSlice {
			stopper := containers.NewContainerStopperPayload(&conSlice[i].ContainerID, nil)
			cController.Start(ctx, stopper).Wait()
			color.PrintStatus("Container", fmt.Sprintf("Stopped %s", conSlice[i].Name))
		}

	})
	color.PrintForInput("Type Ctrl+C to stop running containers\n")

	for {

	}
}

func setupProject(ctx context.Context, projectPath string) {

	deleteProject := func(projectPath string) {
		err := os.RemoveAll(projectPath)
		if err != nil {
			color.PrintFatal(err)
		}
	}

	projectPath = utils.GetAbsChild(projectPath)

	//Deletes project directory upon SIGTERM
	go utils.HandleSIGTERM(func() {
		deleteProject(projectPath)
	})

	projectName := path.Base(projectPath)
	if !utils.IsAlphaNumeric(projectName) {
		color.PrintFatal(fmt.Errorf("Project name should be alpha numeric not %s", projectName))
		deleteProject(projectPath)
	}

	color.PrintStatus("Creating Project", projectName)
	if utils.FileExists(path.Join(projectPath, "persist.db")) {
		color.PrintYellow(fmt.Sprintf("Project %s already exists at %s", projectName, projectPath))
		return
	}
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		deleteProject(projectPath)
		color.PrintFatal(err)
	}
	cController, err := containers.NewController(context.Background(), client)
	if err != nil {
		deleteProject(projectPath)
		color.PrintFatal(err)
	}

	utils.Chdir(utils.MkdirAbs(projectPath))
	//Makes a directory for golang source code generation
	utils.Mkdir("src")
	pController, err := persistence.NewPersistedDataController("persist.db")

	utils.AppendFileIfNotExists(".gitignore", "database/", ".env")
	//creating a variable here to access it in the callback function
	var (
		creater containers.ContainerCreator
	)

	handleCreation := func(ctx context.Context, err error) error {
		if err != nil {
			if containers.IsErrNeedContainerRemove(err) {
				id := containers.GetIDFromNeedContainerRemoveError(err)

				opt := containers.CRMOptions{Force: true, RemoveVolumes: true}

				remover := containers.NewRemoveContainerPayload(id, opt, func(ctx context.Context, err error) error {
					return pController.DeleteContainerByID(id)
				})

				cController.Start(ctx, remover).Wait()
				cController.Start(ctx, creater).Wait()

				return nil
			}

			deleteProject(projectPath)
			return err
		}

		if container, ok := containers.FromContainerContext(ctx); ok {
			//save container to persist.db
			err := pController.Persist(container)
			if err != nil {
				deleteProject(projectPath)
				return err
			}
		}

		return nil
	}

	//Pull default images
	images := []*containers.Image{
		{Name: mongodb.DefaultImgTag},
		{Name: mysql.DefaultImgTag},
		{Name: nginx.DefaultImgTag},
	}
	for i := range images {
		puller := containers.NewImagePullPayload(images[i], containers.DefaultImagePullWriter, nil)
		cController.Start(ctx, puller).Wait()
	}

	if creater, err = nginx.InitialSetup(projectName, handleCreation); err != nil {
		deleteProject(projectPath)
		color.PrintFatal(err)
	}
	//not calling wait here nginx container is created during the database prompt prompt
	cController.Start(ctx, creater)

	switch promptDBType() {
	case MONGO_SEL:
		if creater, err = mongodb.InitialSetup(projectName, handleCreation); err != nil {
			deleteProject(projectPath)
			color.PrintFatal(err)
		}
		cController.Start(ctx, creater).Wait()
	case MYSQL_SEL:
		if creater, err = mysql.InitialSetup(projectName, handleCreation); err != nil {
			deleteProject(projectPath)
			color.PrintFatal(err)
		}
		cController.Start(ctx, creater).Wait()
	}
	color.PrintStatus("Project Created", "Run blah project --start to start developing.")
}

//Prompts user for the database type
func promptDBType() int {
	var db string
	output := "Select a database: \n(1) MongoDB\n(2) Mysql\n: "
	reader := bufio.NewReader(os.Stdin)
	err := utils.GetInput(reader, output, &db, false, "Database Type")
	if err != nil {
		color.PrintFatal(err)
	}
	if db == "1" {
		return MONGO_SEL
	}
	if db == "2" {
		return MYSQL_SEL
	}
	color.PrintYellow("Select 1 for Mongodb or 2 for Mysql.")
	return promptDBType()
}
