package service

import (
	"fmt"
	"log/slog"

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
	slog.Info("AddInventoryItem called", "id", item.IngredientID, "qty", item.Quantity)
	if item.Quantity < 0 {
		slog.Warn("AddInventoryItem: negative quantity", "id", item.IngredientID, "qty", item.Quantity)
		return fmt.Errorf("quantity must be non-negative")
	}

	slog.Info("AddInventoryItem: passing to repo", "id", item.IngredientID)
	err := s.repo.Add(item)
	if err != nil {
		slog.Error("AddInventoryItem: repo.Add failed", "err", err)
		return err
	}
	slog.Info("AddInventoryItem: success", "id", item.IngredientID)
	return nil
}

func (s *inventoryServ) GetAllInventoryItem() ([]models.InventoryItem, error) {
	slog.Info("GetAllInventoryItem called")
	items, err := s.repo.FindAll()
	if err != nil {
		slog.Error("GetAllInventoryItem: repo.FindAll failed", "err", err)
		return nil, err
	}
	slog.Info("GetAllInventoryItem: returning items", "count", len(items))
	return items, nil
}

func (s *inventoryServ) GetInventoryItemByID(id string) (models.InventoryItem, error) {
	slog.Info("GetInventoryItemByID called", "id", id)
	item, err := s.repo.FindByID(id)
	if err != nil {
		slog.Warn("GetInventoryItemByID: not found or repo error", "id", id, "err", err)
		return models.InventoryItem{}, err
	}
	slog.Info("GetInventoryItemByID: found", "id", id)
	return *item, nil
}

func (s *inventoryServ) UpdateInventoryItem(id string, updatedItem models.InventoryItem) error {
	slog.Info("UpdateInventoryItem called", "id", id, "newQty", updatedItem.Quantity)

	if updatedItem.Quantity < 0 {
		slog.Warn("UpdateInventoryItem: negative quantity", "id", id, "qty", updatedItem.Quantity)
		return fmt.Errorf("quantity must be non-negative")
	}

	slog.Info("UpdateInventoryItem: passing to repo", "id", id)
	err := s.repo.Update(id, updatedItem)
	if err != nil {
		slog.Error("UpdateInventoryItem: repo.Update failed", "id", id, "err", err)
		return err
	}
	slog.Info("UpdateInventoryItem: success", "id", id)
	return nil
}

func (s *inventoryServ) DeleteInventoryItem(id string) error {
	slog.Info("DeleteInventoryItem called", "id", id)
	err := s.repo.Delete(id)
	if err != nil {
		slog.Warn("DeleteInventoryItem: repo.Delete failed or not found", "id", id, "err", err)
		return err
	}
	slog.Info("DeleteInventoryItem: success", "id", id)
	return nil
}
