package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"sync"

	"hot-coffee/models"
)

type InventoryRepository interface {
	Add(item models.InventoryItem) error
	FindAll() ([]models.InventoryItem, error)
	FindByID(id string) (*models.InventoryItem, error)
	Update(id string, updated models.InventoryItem) error
	Delete(id string) error
}

type jsonInventoryRepo struct {
	dataDir string
	mu      sync.Mutex
}

func NewJSONInventoryRepo(dir string) InventoryRepository {
	return &jsonInventoryRepo{dataDir: dir}
}

func (r *jsonInventoryRepo) loadInventory() ([]models.InventoryItem, error) {
	path := filepath.Join(r.dataDir, "inventory.json")
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		slog.Error("failed to read inventory file", "path", path, "err", err)
		return nil, err
	}
	var inventory []models.InventoryItem
	if err := json.Unmarshal(raw, &inventory); err != nil {
		slog.Error("failed to unmarshal inventory JSON", "err", err)
		return nil, err
	}

	slog.Info("loadInventory: success", "count", len(inventory))
	return inventory, nil
}

func (r *jsonInventoryRepo) saveInventory(inventory []models.InventoryItem) error {
	raw, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(r.dataDir, "inventory.json")
	slog.Info("saving inventory file", "path", path, "count", len(inventory))
	return ioutil.WriteFile(path, raw, 0o644)
}

func (r *jsonInventoryRepo) Add(item models.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("adding inventory item", "id", item.IngredientID, "name", item.Name)
	inventory, err := r.loadInventory()
	if err != nil {
		slog.Error("Add: loadInventory failed", "err", err)
		return err
	}
	for i := range inventory {
		if inventory[i].IngredientID == item.IngredientID {
			slog.Warn("Add: duplicate ID", "id", item.IngredientID)
			return fmt.Errorf("Item ID already exists")
		}
		if inventory[i].Name == item.Name {
			slog.Warn("Add: duplicate Name", "name", item.Name)
			return fmt.Errorf("Item Name already exists")
		}
	}
	inventory = append(inventory, item)

	if err := r.saveInventory(inventory); err != nil {
		slog.Error("Add: saveInventory failed", "err", err)
		return err
	}
	slog.Info("Add: item added", "id", item.IngredientID)
	return nil
}

func (r *jsonInventoryRepo) FindAll() ([]models.InventoryItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Info("FindAll: called")
	inventory, err := r.loadInventory()
	if err != nil {
		slog.Error("FindAll: loadInventory failed", "err", err)
		return nil, err
	}
	slog.Info("FindAll: returning items", "count", len(inventory))
	return inventory, nil
}

func (r *jsonInventoryRepo) FindByID(id string) (*models.InventoryItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("FindByID: called", "id", id)
	inventory, err := r.loadInventory()
	if err != nil {
		slog.Error("FindByID: loadInventory failed", "err", err)
		return nil, err
	}

	for _, item := range inventory {
		if item.IngredientID == id {
			slog.Info("FindByID: item found", "id", id)
			return &item, nil
		}
	}
	slog.Warn("FindByID: not found", "id", id)
	return nil, fmt.Errorf("inventory item %s not found", id)
}

func (r *jsonInventoryRepo) Update(id string, updated models.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Update: called", "id", id, "newID", updated.IngredientID)
	inventory, err := r.loadInventory()
	if err != nil {
		slog.Error("Update: loadInventory failed", "err", err)
		return err
	}

	if id == updated.IngredientID {
		for i, item := range inventory {
			if item.IngredientID == id {
				inventory[i] = updated
				slog.Info("Update: success", "id", updated.IngredientID)
				return r.saveInventory(inventory)
			}
		}
	} else {
		for _, item := range inventory {
			if item.IngredientID == updated.IngredientID {
				slog.Warn("Update: duplicate ID ", "existing")
				return fmt.Errorf("Inventory item ID already exists")
			}
			if item.Name == updated.Name {
				slog.Warn("Update: duplicate name", "existing")
				return fmt.Errorf("Inventory item name already exists")
			}
		}
		for i, item := range inventory {
			if item.IngredientID == id {
				inventory[i] = updated
				slog.Info("Update: success", "id", updated.IngredientID)
				return r.saveInventory(inventory)
			}
		}
	}

	return fmt.Errorf("inventory item %s not found", id)
}

func (r *jsonInventoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inventory, err := r.loadInventory()
	if err != nil {
		slog.Error("Delete: loadInventory failed", "err", err)
		return err
	}

	filtered := inventory[:0]
	for _, item := range inventory {
		if item.IngredientID != id {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == len(inventory) {
		slog.Warn("Delete: item not found", "id", id)
		return fmt.Errorf("inventory item %s not found", id)
	}

	if err := r.saveInventory(filtered); err != nil {
		slog.Error("Delete: saveInventory failed", "err", err)
		return err
	}
	slog.Info("Delete: success", "id", id)
	return nil
}
