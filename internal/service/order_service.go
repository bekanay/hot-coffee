package service

import (
	"fmt"
	"hot-coffee/internal/repository"
	"hot-coffee/models"
)

type OrderService interface {
	CreateOrder(order models.Order) error
	GetOrders() ([]models.Order, error)
	GetOrderById(id string) (models.Order, error)
	UpdateOrder(id string, order models.Order) error
	DeleteOrder(id string) error
	CloseOrder(id string) error
	GetTotalSales() (models.Total, error)
	GetPopularMenuItems() ([]models.MenuItem, error)
}

type OrderServ struct {
	orderRepo repository.OrderRepository
	menuRepo  repository.MenuRepository
	invRepo   repository.InventoryRepository
}

func NewOrderService(or repository.OrderRepository, mr repository.MenuRepository, ir repository.InventoryRepository) *OrderServ {
	return &OrderServ{orderRepo: or, menuRepo: mr, invRepo: ir}
}

func (s *OrderServ) CreateOrder(order models.Order) error {
	err := checkForIngredients(s, order)
	if err != nil {
		return err
	}
	err = s.orderRepo.Add(order)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderServ) GetOrders() ([]models.Order, error) {
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderServ) GetOrderById(id string) (models.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return models.Order{}, err
	}

	return *order, nil
}

func (s *OrderServ) UpdateOrder(id string, updatedOrder models.Order) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}

	var missingMenuItems []models.OrderItem

	for _, menuItem := range order.Items {
		exists := false
		for _, menuUpdatedItem := range updatedOrder.Items {
			if menuItem.ProductID == menuUpdatedItem.ProductID {
				exists = true
			}
		}
		if !exists {
			missingMenuItems = append(missingMenuItems, menuItem)
		}
	}

	err = checkForIngredients(s, updatedOrder)
	if err != nil {
		return err
	}

	err = s.orderRepo.Update(id, updatedOrder)
	if err != nil {
		return err
	}

	return nil
}

func checkForIngredients(s *OrderServ, order models.Order) error {
	invItems, err := s.invRepo.FindAll()
	if err != nil {
		return err
	}

	for _, menuItem := range order.Items {
		item, err := s.menuRepo.FindByID(menuItem.ProductID)
		if err != nil {
			return err
		}
		for _, ingredient := range item.Ingredients {
			for _, invItem := range invItems {
				if invItem.IngredientID == ingredient.IngredientID {
					if ingredient.Quantity*float64(menuItem.Quantity) > invItem.Quantity {
						return fmt.Errorf("not enough ingredient %s", ingredient.IngredientID)
					}
					invItem.Quantity -= ingredient.Quantity * float64(menuItem.Quantity)
				}
			}
		}
	}
	return nil
}

func (s *OrderServ) DeleteOrder(id string) error {
	err := s.orderRepo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderServ) CloseOrder(id string) error {
	err := s.orderRepo.Close(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderServ) GetTotalSales() (models.Total, error) {
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		return models.Total{}, err
	}

	totalSales := 0.0
	for _, order := range orders {
		for _, menuItem := range order.Items {
			if order.Status == "closed" {
				item, err := s.menuRepo.FindByID(menuItem.ProductID)
				if err != nil {
					return models.Total{}, err
				}
				totalSales += item.Price * float64(menuItem.Quantity)
			}
		}
	}
	return models.Total{TotalSales: totalSales}, nil
}

func (s *OrderServ) GetPopularMenuItems() ([]models.MenuItem, error) {
	return []models.MenuItem{}, nil
}
