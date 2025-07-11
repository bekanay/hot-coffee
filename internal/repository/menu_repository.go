package repository

import (
	"encoding/json"
	"fmt"
	"hot-coffee/models"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type MenuRepository interface {
	Add(item models.MenuItem) error
	FindAll() ([]models.MenuItem, error)
	FindByID(id string) (*models.MenuItem, error)
	Update(id string, updated models.MenuItem) error
	Delete(id string) error
}

type jsonMenuRepo struct {
	dataDir string
	mu      sync.Mutex
}

func NewJSONMenuRepo(dir string) MenuRepository {
	return &jsonMenuRepo{dataDir: dir}
}

func (r *jsonMenuRepo) loadMenuItems() ([]models.MenuItem, error) {
	path := filepath.Join(r.dataDir, "menu_items.json")
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var menuItems []models.MenuItem
	if err := json.Unmarshal(raw, &menuItems); err != nil {
		return nil, err
	}

	return menuItems, nil
}

func (r *jsonMenuRepo) saveMenuItems(menuItems []models.MenuItem) error {
	raw, err := json.MarshalIndent(menuItems, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(r.dataDir, "menu_items.json")
	return ioutil.WriteFile(path, raw, 0644)
}

func (r *jsonMenuRepo) Add(menuItem models.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	menuItems, err := r.loadMenuItems()
	if err != nil {
		return err
	}
	menuItems = append(menuItems, menuItem)

	return r.saveMenuItems(menuItems)
}

func (r *jsonMenuRepo) FindAll() ([]models.MenuItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.loadMenuItems()
}

func (r *jsonMenuRepo) FindByID(id string) (*models.MenuItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	menuItems, err := r.loadMenuItems()
	if err != nil {
		return nil, err
	}

	for _, item := range menuItems {
		if item.ID == id {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("menu item %s not found", id)
}

func (r *jsonMenuRepo) Update(id string, updated models.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	menuItems, err := r.loadMenuItems()
	if err != nil {
		return err
	}

	for i, item := range menuItems {
		if item.ID == id {
			menuItems[i] = updated
			return r.saveMenuItems(menuItems)
		}
	}

	return fmt.Errorf("menu item %s not found", id)
}

func (r *jsonMenuRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	menuItems, err := r.loadMenuItems()
	if err != nil {
		return err
	}

	filtered := menuItems[:0]
	for _, item := range menuItems {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == len(menuItems) {
		return fmt.Errorf("menu item %s not found", id)
	}

	return r.saveMenuItems(filtered)
}
