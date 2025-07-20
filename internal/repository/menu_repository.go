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
	slog.Info("loadMenuItems", "path", path)

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		slog.Error("loadMenuItems: ReadFile failed", "path", path, "err", err)
		return nil, err
	}
	var menuItems []models.MenuItem
	if err := json.Unmarshal(raw, &menuItems); err != nil {
		slog.Error("loadMenuItems: Unmarshal failed", "err", err)
		return nil, err
	}

	slog.Info("loadMenuItems: success", "count", len(menuItems))
	return menuItems, nil
}

func (r *jsonMenuRepo) saveMenuItems(menuItems []models.MenuItem) error {
	slog.Info("saveMenuItems: called", "count", len(menuItems))
	raw, err := json.MarshalIndent(menuItems, "", "  ")
	if err != nil {
		slog.Error("saveMenuItems: MarshalIndent failed", "err", err)
		return err
	}
	path := filepath.Join(r.dataDir, "menu_items.json")
	if err := ioutil.WriteFile(path, raw, 0o644); err != nil {
		slog.Error("saveMenuItems: WriteFile failed", "path", path, "err", err)
		return err
	}
	slog.Info("saveMenuItems: success", "path", path)
	return nil
}

func (r *jsonMenuRepo) Add(menuItem models.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Info("Add: called", "id", menuItem.ID, "name", menuItem.Name)

	menuItems, err := r.loadMenuItems()
	if err != nil {
		slog.Error("Add: loadMenuItems failed", "err", err)
		return err
	}
	for _, item := range menuItems {
		if item.ID == menuItem.ID {
			slog.Warn("Add: duplicate ID", "id", menuItem.ID)
			return fmt.Errorf("Menu item ID already exists")
		}
		if item.Name == menuItem.Name {
			slog.Warn("Add: duplicate Name", "name", menuItem.Name)
			return fmt.Errorf("Menu item name already exists")
		}
	}
	slog.Info("Add: appending item")
	menuItems = append(menuItems, menuItem)

	if err := r.saveMenuItems(menuItems); err != nil {
		slog.Error("Add: saveMenuItems failed", "err", err)
		return err
	}
	slog.Info("Add: success", "id", menuItem.ID)
	return nil
}

func (r *jsonMenuRepo) FindAll() ([]models.MenuItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Info("FindAll: called")

	items, err := r.loadMenuItems()
	if err != nil {
		slog.Error("FindAll: loadMenuItems failed", "err", err)
		return nil, err
	}
	slog.Info("FindAll: returning items", "count", len(items))
	return items, nil
}

func (r *jsonMenuRepo) FindByID(id string) (*models.MenuItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("FindByID: called", "id", id)
	menuItems, err := r.loadMenuItems()
	if err != nil {
		slog.Error("FindByID: loadMenuItems failed", "err", err)
		return nil, err
	}

	for _, item := range menuItems {
		if item.ID == id {
			slog.Info("FindByID: found", "id", id)
			return &item, nil
		}
	}
	slog.Warn("FindByID: not found", "id", id)
	return nil, fmt.Errorf("menu item %s not found", id)
}

func (r *jsonMenuRepo) Update(id string, updated models.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Info("Update: called", "id", id)

	menuItems, err := r.loadMenuItems()
	if err != nil {
		slog.Error("Update: loadMenuItems failed", "err", err)
		return err
	}

	for i, item := range menuItems {
		if item.ID == id {
			slog.Info("Update: applying update", "index", i)
			menuItems[i] = updated
			if err := r.saveMenuItems(menuItems); err != nil {
				slog.Error("Update: saveMenuItems failed", "err", err)
			}
			slog.Info("Update: success", "id", id)
			return err
		}
		if item.ID == updated.ID {
			slog.Warn("Update: duplicate ID", "conflictID", updated.ID)
			return fmt.Errorf("Menu item ID already exists")
		}
		if item.Name == updated.Name {
			slog.Warn("Update: duplicate Name", "conflictName", updated.Name)
			return fmt.Errorf("Menu item name already exists")
		}
	}

	slog.Warn("Update: not found", "id", id)
	return fmt.Errorf("menu item %s not found", id)
}

func (r *jsonMenuRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Delete: called", "id", id)
	menuItems, err := r.loadMenuItems()
	if err != nil {
		slog.Error("Delete: loadMenuItems failed", "err", err)
		return err
	}

	filtered := menuItems[:0]
	for _, item := range menuItems {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == len(menuItems) {
		slog.Warn("Delete: not found", "id", id)
		return fmt.Errorf("menu item %s not found", id)
	}

	if err := r.saveMenuItems(filtered); err != nil {
		slog.Error("Delete: saveMenuItems failed", "err", err)
		return err
	}
	slog.Info("Delete: success", "id", id)
	return nil
}
