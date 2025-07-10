package repository

import "hot-coffee/models"

type InventoryRepository interface {
	Add(item models.InventoryItem) error
	FindAll() ([]models.InventoryItem, error)
	FindById(id string) (models.InventoryItem, error)
	Update(id string) error
	Delete(id string) error
}

type inventoryRepository struct{}
