package repository

import "hot-coffee/models"

type MenuRepository interface {
	Add(item models.MenuItem) error
	FindAll() ([]models.MenuItem, error)
	FindById(id string) (models.MenuItem, error)
	Update(id string) error
	Delete(id string) error
}

type menuRepository struct{}
