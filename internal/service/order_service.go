package service

import (
	"fmt"
	"hot-coffee/internal/repository"
	"hot-coffee/models"
	"strconv"
	"time"
)

type OrderService interface {
	CreateOrder(order models.Order) ([]string, error)
	GetOrders() ([]models.Order, error)
	GetOrderById(id string) (models.Order, error)
	UpdateOrder(id string, order models.Order) ([]string, error)
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

func (s *OrderServ) CreateOrder(order models.Order) ([]string, error) {
	validateConflicts, err := validateOrder(s, order)
	if err != nil {
		return validateConflicts, err
	}

	if len(validateConflicts) != 0 {
		return validateConflicts, nil
	}

	for _, product := range order.Items {
		if product.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity %d for product: %s", product.Quantity, product.ProductID)
		}
	}

	requiredIngredients, err := countRequired(s, order)
	if err != nil {
		return nil, err
	}

	conflicts, err := compareIngredients(s, order, requiredIngredients)
	if err != nil {
		return conflicts, err
	}

	if len(conflicts) != 0 {
		return conflicts, nil
	} else {
		err := orderResult(s, requiredIngredients)
		if err != nil {
			return nil, err
		}
	}

	if order.Status == "" {
		order.Status = "open"
	}

	if order.CreatedAt == "" {
		order.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	err = s.orderRepo.Add(order)
	if err != nil {
		return nil, err
	}

	return nil, nil
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

func (s *OrderServ) UpdateOrder(id string, updatedOrder models.Order) ([]string, error) {
	for _, product := range updatedOrder.Items {
		if product.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity %d for product: %s", product.Quantity, product.ProductID)
		}
	}

	// order, err := s.orderRepo.FindByID(id)
	// if err != nil {
	// 	return nil, err
	// }

	requiredIngredients, err := countRequired(s, updatedOrder)
	if err != nil {
		return nil, err
	}

	conflicts, err := compareIngredients(s, updatedOrder, requiredIngredients)
	if err != nil {
		return conflicts, err
	}

	if len(conflicts) != 0 {
		return conflicts, nil
	} else {
		err := orderResult(s, requiredIngredients)
		if err != nil {
			return nil, err
		}
	}

	// var missingMenuItems []models.OrderItem

	// for _, menuItem := range order.Items {
	// 	exists := false
	// 	for _, menuUpdatedItem := range updatedOrder.Items {
	// 		if menuItem.ProductID == menuUpdatedItem.ProductID {
	// 			exists = true
	// 		}
	// 	}

	// 	if !exists {
	// 		missingMenuItems = append(missingMenuItems, menuItem)
	// 	}
	// }

	// err = checkForIngredients(s, updatedOrder)
	// if err != nil {
	// 	return err
	// }

	err = s.orderRepo.Update(id, updatedOrder)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func validateOrder(s *OrderServ, order models.Order) ([]string, error) {
	var validateConflicts []string

	_, err := s.orderRepo.FindByID(order.ID)
	if err == nil {
		validateConflicts = append(validateConflicts, models.ErrAlreadyExists.Error())
	}

	if order.ID == "" {
		validateConflicts = append(validateConflicts, "order id is empty")
	}

	if order.CustomerName == "" {
		validateConflicts = append(validateConflicts, "customer name is empty")
	}

	if len(order.Items) == 0 {
		validateConflicts = append(validateConflicts, "order items is empty")
	}
	return validateConflicts, nil
}

func countRequired(s *OrderServ, order models.Order) (map[string]float64, error) {
	requiredIngredients := make(map[string]float64)
	for _, menuItem := range order.Items {
		if menuItem.Quantity < 0 {
			return nil, fmt.Errorf("menu item %s has negative value", menuItem.ProductID)
		}

		item, err := s.menuRepo.FindByID(menuItem.ProductID)
		if err != nil {
			return nil, err
		}

		for _, ingredient := range item.Ingredients {
			if ingredient.Quantity < 0 {
				requiredIngredients[ingredient.IngredientID] = -1
				// return nil, fmt.Errorf("ingredient %s has negative value", ingredient.IngredientID)
			}
			if _, err := s.invRepo.FindByID(ingredient.IngredientID); err == nil {
				if _, ok := requiredIngredients[ingredient.IngredientID]; ok {
					requiredIngredients[ingredient.IngredientID] += ingredient.Quantity * float64(menuItem.Quantity)
				} else {
					requiredIngredients[ingredient.IngredientID] = ingredient.Quantity * float64(menuItem.Quantity)
				}
			} else {
				requiredIngredients[ingredient.IngredientID] = -2
				// return nil, fmt.Errorf("ingredient %s not found", ingredient.IngredientID)
			}
		}
	}
	return requiredIngredients, nil
}

func compareIngredients(s *OrderServ, order models.Order, requiredIngredients map[string]float64) ([]string, error) {
	var conflicts []string

	for key, val := range requiredIngredients {
		var temp string
		if val == -1 {
			temp = "ingredient " + key + " has negative value"
			conflicts = append(conflicts, temp)
			continue
		}

		if val == -2 {
			temp = "ingredient " + key + " not found"
			conflicts = append(conflicts, temp)
			continue
		}

		invItem, err := s.invRepo.FindByID(key)
		if err != nil {
			return conflicts, err
		}
		if val > invItem.Quantity {
			temp = "Insufficient inventory for ingredient '" + key + "'. Required: " + strconv.Itoa(int(val)) + invItem.Unit + ", Available: " + strconv.Itoa(int(invItem.Quantity)) + invItem.Unit
			conflicts = append(conflicts, temp)
		}
	}
	return conflicts, nil
}

func orderResult(s *OrderServ, requiredIngredients map[string]float64) error {
	invItems, err := s.invRepo.FindAll()
	if err != nil {
		return err
	}
	for key, val := range requiredIngredients {
		for i := range invItems {
			if invItems[i].IngredientID == key {
				invItems[i].Quantity -= val
				if err := s.invRepo.Update(key, invItems[i]); err != nil {
					return nil
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
	// orders, err := s.orderRepo.FindAll()
	// if err != nil {
	// 	return nil, err
	// }

	// counter := 0
	// orderMap := make(map[string]int)
	// for _, order := range orders {
	// 	if order.ID
	// }
	return []models.MenuItem{}, nil
}
