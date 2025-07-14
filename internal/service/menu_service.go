package service

import (
	"fmt"
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
	if item.Price <= 0 {
		return fmt.Errorf("price must be non-negative")
	}

	for _, ingredient := range item.Ingredients {
		if ingredient.Quantity < 0 {
			return fmt.Errorf("quantity must be non-negative value")
		}
	}
	menuItems, _ := s.menuRepo.FindAll()
	for _, exist := range menuItems {
		if exist.ID == item.ID {
			return fmt.Errorf("Menu item ID already exists")
		}
		if exist.Name == item.Name {
			return fmt.Errorf("Menu item name already exists")
		}

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
				if found == false {
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
				return fmt.Errorf("following ingredients missing: %s", list)
			}
		}
	}

	return s.menuRepo.Add(item)
}

func (s *menuServ) GetAllMenuItems() ([]models.MenuItem, error) {
	return s.menuRepo.FindAll()
}

func (s *menuServ) GetMenuItemByID(id string) (models.MenuItem, error) {
	item, err := s.menuRepo.FindByID(id)
	if err != nil {
		return models.MenuItem{}, err
	}

	return *item, nil
}

func (s *menuServ) UpdateMenuItem(id string, updatedItem models.MenuItem) error {
	if updatedItem.Price <= 0 {
		return fmt.Errorf("price must be non-negative")
	}

	return s.menuRepo.Update(id, updatedItem)
}

func (s *menuServ) DeleteMenuItem(id string) error {
	return s.menuRepo.Delete(id)
}
