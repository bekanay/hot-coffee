package service

import "hot-coffee/models"

type MenuService interface {
	AddMenuItem(item models.MenuItem) error
	GetAllMenuItems() ([]models.MenuItem, error)
	GetMenuItemByID(id string) (models.MenuItem, error)
	UpdateMenuItem(id string, updatedItem models.MenuItem) error
	DeleteMenuItem(id string) error
}
