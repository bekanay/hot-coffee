package service

import (
	"hot-coffee/models"
)

type OrderService interface {
	CreateOrder(models.Order) (models.Order, error)
	GetOrders() ([]models.Order, error)
	GetOrderById(id string) (models.Order, error)
	UpdateOrder(id string, order models.Order) error
	DeleteOrder(id string) error
	CloseOrder(id string) error
	GetTotalSales() (float64, error)
	GetPopularMenuItems() ([]models.MenuItem, error)
}
