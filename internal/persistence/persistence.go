package persistence

import (
	"github.com/dannyvidal/blah/internal/containers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PersistedDataController struct{ db *gorm.DB }

func (c *PersistedDataController) Persist(object any) error {
	return c.db.Save(object).Error
}

func (c *PersistedDataController) GetAllContainers() ([]*containers.Container, error) {
	var containers []*containers.Container
	tx := c.db.Find(&containers)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return containers, nil
}

func (c *PersistedDataController) FindOneContainerByID(ID string) (*containers.Container, error) {
	var container containers.Container
	tx := c.db.Where("ContainerID LIKE ?", "%", ID, "%").First(&container)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &container, nil
}

//Deletes a persisted container from sqlite file along with its assosiated mount points
func (c *PersistedDataController) DeleteContainerByID(ID string) error {

	tx := c.db.Unscoped().Select("Mounts").Where("container_id = ?", ID).Delete(&containers.Container{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

//Opens a sqlite file and returns a controller to manage persistence
func NewPersistedDataController(name string) (*PersistedDataController, error) {
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})

	db.AutoMigrate(&containers.Container{})
	db.AutoMigrate(&containers.ContainerMount{})
	db.AutoMigrate(&containers.ContainerExposedPort{})
	db.AutoMigrate(&containers.ContainerPortBinding{})
	db.AutoMigrate(&containers.ContainerEnv{})

	if err != nil {
		return nil, err
	}
	return &PersistedDataController{db}, nil
}
