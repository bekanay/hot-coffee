package service

import (
	"fmt"
	"log/slog"

	"hot-coffee/internal/repository"
	"hot-coffee/models"
)

type MenuService interface {
	AddMenuItem(item models.MenuItem) error
	GetAllMenuItems() ([]models.MenuItem, error)
	GetMenuItemByID(id string) (models.MenuItem, error)
	UpdateMenuItem(id string, updatedItem models.MenuItem) error
	DeleteMenuItem(id string) error
}

type menuServ struct {
	menuRepo repository.MenuRepository
	invRepo  repository.InventoryRepository
}

func NewMenuService(mr repository.MenuRepository, ir repository.InventoryRepository) MenuService {
	return &menuServ{menuRepo: mr, invRepo: ir}
}

func (s *menuServ) AddMenuItem(item models.MenuItem) error {
	slog.Info("AddMenuItem called", "id", item.ID, "name", item.Name, "price", item.Price)

	if item.Price <= 0 {
		slog.Warn("AddMenuItem: non-positive price", "price", item.Price)
		return fmt.Errorf("price must be non-negative")
	}

	for _, ingredient := range item.Ingredients {
		if ingredient.Quantity < 0 {
			slog.Warn("AddMenuItem: ingredient negative quantity", "ingredient", ingredient.IngredientID, "qty", ingredient.Quantity)
			return fmt.Errorf("quantity must be non-negative value")
		}
	}

	slog.Info("AddMenuItem: loading existing menu items")
	menuItems, _ := s.menuRepo.FindAll()

	for _, exist := range menuItems {
		if exist.ID == item.ID {
			slog.Warn("AddMenuItem: duplicate ID", "id", item.ID)
			return fmt.Errorf("Menu item ID already exists")
		}
		if exist.Name == item.Name {
			slog.Warn("AddMenuItem: duplicate name", "name", item.Name)
			return fmt.Errorf("Menu item name already exists")
		}

		slog.Info("AddMenuItem: checking inventory for ingredients")
		inventoryItems, _ := s.invRepo.FindAll()
		var matchedIngredients []string
		for _, invItem := range inventoryItems {
			for _, ingredient := range item.Ingredients {
				if invItem.IngredientID == ingredient.IngredientID {
					matchedIngredients = append(matchedIngredients, ingredient.IngredientID)
				}
			}
		}
		var missingIngredients []string
		if len(matchedIngredients) != len(item.Ingredients) {
			for _, ingredient := range item.Ingredients {
				found := false
				for _, mIngredient := range matchedIngredients {
					if ingredient.IngredientID == mIngredient {
						found = true
					}
				}
				if !found {
					missingIngredients = append(missingIngredients, ingredient.IngredientID)
				}
			}
			if len(missingIngredients) != 0 {
				var list string
				for i, ingredient := range missingIngredients {
					list += ingredient
					if i != len(missingIngredients)-1 {
						list += ", "
					}
				}
				slog.Warn("AddMenuItem: missing ingredients", "missing", missingIngredients)
				return fmt.Errorf("following ingredients missing: %s", list)
			}
		}
	}

	slog.Info("AddMenuItem: saving new menu item to repo")
	err := s.menuRepo.Add(item)
	if err != nil {
		slog.Error("AddMenuItem: repo.Add failed", "err", err)
		return err
	}
	slog.Info("AddMenuItem: success", "id", item.ID)
	return nil
}

func (s *menuServ) GetAllMenuItems() ([]models.MenuItem, error) {
	slog.Info("GetAllMenuItems called")
	items, err := s.menuRepo.FindAll()
	if err != nil {
		slog.Error("GetAllMenuItems: repo.FindAll failed", "err", err)
		return nil, err
	}
	slog.Info("GetAllMenuItems: returning items", "count", len(items))
	return items, nil
}

func (s *menuServ) GetMenuItemByID(id string) (models.MenuItem, error) {
	slog.Info("GetMenuItemByID called", "id", id)
	ptr, err := s.menuRepo.FindByID(id)
	if err != nil {
		slog.Warn("GetMenuItemByID: not found", "id", id, "err", err)
		return models.MenuItem{}, err
	}
	slog.Info("GetMenuItemByID: found", "id", id)
	return *ptr, nil
}

func (s *menuServ) UpdateMenuItem(id string, updatedItem models.MenuItem) error {
	slog.Info("UpdateMenuItem called", "id", id)
	if updatedItem.Price <= 0 {
		slog.Warn("UpdateMenuItem: non-positive price", "price", updatedItem.Price)
		return fmt.Errorf("price must be non-negative")
	}
	slog.Info("UpdateMenuItem: passing update to repo", "id", id)
	err := s.menuRepo.Update(id, updatedItem)
	if err != nil {
		slog.Error("UpdateMenuItem: repo.Update failed", "id", id, "err", err)
		return err
	}
	slog.Info("UpdateMenuItem: success", "id", id)
	return nil
}

func (s *menuServ) DeleteMenuItem(id string) error {
	slog.Info("DeleteMenuItem called", "id", id)
	err := s.menuRepo.Delete(id)
	if err != nil {
		slog.Warn("DeleteMenuItem: repo.Delete failed", "id", id, "err", err)
		return err
	}
	slog.Info("DeleteMenuItem: success", "id", id)
	return nil
}
