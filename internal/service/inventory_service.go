package service

import (
	"fmt"
	"hot-coffee/internal/repository"
	"hot-coffee/models"
)

type InventoryService interface {
	AddInventoryItem(item models.InventoryItem) error
	GetAllInventoryItem() ([]models.InventoryItem, error)
	GetInventoryItemByID(id string) (models.InventoryItem, error)
	UpdateInventoryItem(id string, updatedItem models.InventoryItem) error
	DeleteInventoryItem(id string) error
}

type inventoryServ struct {
	repo repository.InventoryRepository
}

func NewInventoryService(r repository.InventoryRepository) InventoryService {
	return &inventoryServ{repo: r}
}

func (s *inventoryServ) AddInventoryItem(item models.InventoryItem) error {
	if item.Quantity < 0 {
		return fmt.Errorf("quantity must be non-negative")
	}

	return s.repo.Add(item)
}

func (s *inventoryServ) GetAllInventoryItem() ([]models.InventoryItem, error) {
	return s.repo.FindAll()
}

func (s *inventoryServ) GetInventoryItemByID(id string) (models.InventoryItem, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return models.InventoryItem{}, err
	}
	return *item, nil
}

func (s *inventoryServ) UpdateInventoryItem(id string, updatedItem models.InventoryItem) error {
	if updatedItem.Quantity < 0 {
		return fmt.Errorf("quantity must be non-negative")
	}

	return s.repo.Update(id, updatedItem)
}

func (s *inventoryServ) DeleteInventoryItem(id string) error {
	return s.repo.Delete(id)
}
