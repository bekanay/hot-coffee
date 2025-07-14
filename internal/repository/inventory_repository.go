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
		return nil, err
	}

	return inventory, nil
}

func (r *jsonInventoryRepo) saveInventory(inventory []models.InventoryItem) error {
	raw, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(r.dataDir, "inventory.json")
	return ioutil.WriteFile(path, raw, 0o644)
}

func (r *jsonInventoryRepo) Add(item models.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inventory, err := r.loadInventory()
	if err != nil {
		return err
	}
	for i := range inventory {
		if inventory[i].IngredientID == item.IngredientID {
			return fmt.Errorf("Item ID already exists")
		}
		if inventory[i].Name == item.Name {
			return fmt.Errorf("Item Name already exists")
		}
	}
	inventory = append(inventory, item)

	return r.saveInventory(inventory)
}

func (r *jsonInventoryRepo) FindAll() ([]models.InventoryItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.loadInventory()
}

func (r *jsonInventoryRepo) FindByID(id string) (*models.InventoryItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inventory, err := r.loadInventory()
	if err != nil {
		return nil, err
	}

	for _, item := range inventory {
		if item.IngredientID == id {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("inventory item %s not found", id)
}

func (r *jsonInventoryRepo) Update(id string, updated models.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inventory, err := r.loadInventory()
	if err != nil {
		return err
	}

	for i, item := range inventory {
		if item.IngredientID == id {
			inventory[i] = updated
			return r.saveInventory(inventory)
		}

		if item.IngredientID == updated.IngredientID || item.Name == updated.Name {
			return fmt.Errorf("Inventory item ID or name already exists")
		}
	}

	return fmt.Errorf("inventory item %s not found", id)
}

func (r *jsonInventoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inventory, err := r.loadInventory()
	if err != nil {
		return err
	}

	filtered := inventory[:0]
	for _, item := range inventory {
		if item.IngredientID != id {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == len(inventory) {
		return fmt.Errorf("inventory item %s not found", id)
	}

	return r.saveInventory(filtered)
}
