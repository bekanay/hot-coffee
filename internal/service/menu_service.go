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
	repo repository.MenuRepository
}

func NewMenuService(r repository.MenuRepository) MenuService {
	return &menuServ{repo: r}
}

func (s *menuServ) AddMenuItem(item models.MenuItem) error {
	if item.Price <= 0 {
		return fmt.Errorf("price must be non-negative")
	}

	return s.repo.Add(item)
}

func (s *menuServ) GetAllMenuItems() ([]models.MenuItem, error) {
	return s.repo.FindAll()
}

func (s *menuServ) GetMenuItemByID(id string) (models.MenuItem, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return models.MenuItem{}, err
	}

	return *item, nil
}

func (s *menuServ) UpdateMenuItem(id string, updatedItem models.MenuItem) error {
	if updatedItem.Price <= 0 {
		return fmt.Errorf("price must be non-negative")
	}

	return s.repo.Update(id, updatedItem)
}

func (s *menuServ) DeleteMenuItem(id string) error {
	return s.repo.Delete(id)
}
