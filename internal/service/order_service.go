package service

import (
	"fmt"
	"hot-coffee/internal/repository"
	"hot-coffee/models"
	"strconv"
	"strings"
	"unicode"
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
	for _, product := range order.Items {
		if product.Quantity <= 0 {
			return fmt.Errorf("invalid quantity %d for product: %s", product.Quantity, product.ProductID)
		}
	}
	_, err := s.orderRepo.FindByID(order.ID)
	if err == nil {
		return models.ErrAlreadyExists
	}

	err = checkForIngredients(s, order)
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

	var reducedItems []models.InventoryItem
	noIngredient := false
	requiredIngredients := make(map[string]string)

	for _, menuItem := range order.Items {
		if menuItem.Quantity < 0 {
			return fmt.Errorf("menu item %s has negative value", menuItem.ProductID)
		}
		item, err := s.menuRepo.FindByID(menuItem.ProductID)
		if err != nil {
			return err
		}

		for _, ingredient := range item.Ingredients {
			if ingredient.Quantity < 0 {
				return fmt.Errorf("ingredient %s has negative value", ingredient.IngredientID)
			}
			ingredientFound := false
			for i, invItem := range invItems {
				if invItems[i].IngredientID == ingredient.IngredientID {
					ingredientFound = true
					_, ok := requiredIngredients[ingredient.IngredientID]
					if ok {
						requiredIngredients[ingredient.IngredientID] = strings.Trim(requiredIngredients[ingredient.IngredientID], invItems[i].Unit)
						temp, err := strconv.Atoi(requiredIngredients[ingredient.IngredientID])
						if err != nil {
							return err
						}
						temp += int(ingredient.Quantity) * menuItem.Quantity
						requiredIngredients[ingredient.IngredientID] = strconv.Itoa(temp) + invItems[i].Unit
						temp = 0
					} else {
						requiredIngredients[ingredient.IngredientID] = strconv.Itoa(int(ingredient.Quantity)*menuItem.Quantity) + invItems[i].Unit
					}
					if ingredient.Quantity*float64(menuItem.Quantity) > invItems[i].Quantity {
						noIngredient = true
					} else {
						reducedQuantity := ingredient.Quantity * float64(menuItem.Quantity)
						invItems[i].Quantity -= reducedQuantity
						s.invRepo.Update(invItems[i].IngredientID, invItems[i])
						invItem.Quantity = reducedQuantity
						reducedItems = append(reducedItems, invItem)
					}
				}
			}
			if !ingredientFound {
				returnItems(s, reducedItems)
				return fmt.Errorf("no ingredient found %s", ingredient.IngredientID)
			}
		}
	}

	if noIngredient {
		if err := returnItems(s, reducedItems); err != nil {
			return err
		}

		list := ""
		for key, val := range requiredIngredients {
			number := strings.TrimRightFunc(val, unicode.IsLetter)

			req, err := strconv.Atoi(number)
			if err != nil {
				return err
			}

			ingredient, err := s.invRepo.FindByID(key)
			if err != nil {
				return err
			}

			avail := int(ingredient.Quantity)
			if avail < req {
				list += key + ". Required: " + val + " , Available: " + strconv.Itoa(int(ingredient.Quantity)) + ingredient.Unit + "."
			}
		}

		return fmt.Errorf("Insufficient inventory for ingredient: " + list)
	}

	return nil
}

func returnItems(s *OrderServ, reducedItems []models.InventoryItem) error {
	invItems, err := s.invRepo.FindAll()
	if err != nil {
		return err
	}

	for i := range invItems {
		for _, reducedItem := range reducedItems {
			if invItems[i].IngredientID == reducedItem.IngredientID {
				invItems[i].Quantity += reducedItem.Quantity
				s.invRepo.Update(invItems[i].IngredientID, invItems[i])
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
