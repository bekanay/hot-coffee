package service

import (
	"fmt"
	"hot-coffee/internal/repository"
	"hot-coffee/models"
	"log/slog"
	"sort"
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
	GetPopularMenuItems() ([]string, error)
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
	slog.Info("CreateOrder", slog.String("order_id", order.ID), slog.String("customer", order.CustomerName))
	validateConflicts, err := validateOrder(s, order)
	if err != nil {
		slog.Error("validateOrder", slog.Any("error", err))
		return validateConflicts, err
	}
	if len(validateConflicts) != 0 {
		slog.Warn("validation conflicts", slog.String("order_id", order.ID), slog.Any("conflicts", validateConflicts))
		return validateConflicts, nil
	}
	for _, product := range order.Items {
		if product.Quantity <= 0 {
			slog.Error("invalid quantity", slog.String("product", product.ProductID), slog.Int("quantity", product.Quantity))
			return nil, fmt.Errorf("invalid quantity %d for product: %s", product.Quantity, product.ProductID)
		}
	}
	requiredIngredients, err := countRequired(s, order)
	if err != nil {
		slog.Error("countRequired", slog.Any("error", err))
		return nil, err
	}
	conflicts, err := compareIngredients(s, order, requiredIngredients)
	if err != nil {
		slog.Error("compareIngredients", slog.Any("error", err))
		return conflicts, err
	}
	if len(conflicts) != 0 {
		slog.Warn("ingredient conflicts", slog.String("order_id", order.ID), slog.Any("conflicts", conflicts))
		return conflicts, nil
	}
	if err := orderResult(s, requiredIngredients); err != nil {
		slog.Error("orderResult", slog.Any("error", err))
		return nil, err
	}
	if order.Status == "" {
		order.Status = "open"
	}
	if order.CreatedAt == "" {
		order.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if err := s.orderRepo.Add(order); err != nil {
		slog.Error("Add order", slog.Any("error", err))
		return nil, err
	}
	slog.Info("Order created", slog.String("order_id", order.ID))
	return nil, nil
}

func (s *OrderServ) GetOrders() ([]models.Order, error) {
	slog.Info("GetOrders")
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		slog.Error("FindAll", slog.Any("error", err))
		return nil, err
	}
	slog.Info("GetOrders result", slog.Int("count", len(orders)))
	return orders, nil
}

func (s *OrderServ) GetOrderById(id string) (models.Order, error) {
	slog.Info("GetOrderById", slog.String("order_id", id))
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		slog.Error("FindByID", slog.String("order_id", id), slog.Any("error", err))
		return models.Order{}, err
	}
	return *order, nil
}

func (s *OrderServ) UpdateOrder(id string, updatedOrder models.Order) ([]string, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		slog.Error("FindByID", slog.String("order_id", id), slog.Any("error", err))
		return nil, err
	}

	if order.ID != updatedOrder.ID {
		_, err = s.orderRepo.FindByID(updatedOrder.ID)
		if err == nil {
			return nil, fmt.Errorf("order with this id already exists")
		}
	}

	slog.Info("UpdateOrder", slog.String("order_id", id))
	for _, product := range updatedOrder.Items {
		if product.Quantity <= 0 {
			slog.Error("invalid quantity", slog.String("product", product.ProductID), slog.Int("quantity", product.Quantity))
			return nil, fmt.Errorf("invalid quantity %d for product: %s", product.Quantity, product.ProductID)
		}
	}

	requiredIngredientsPrev, err := countRequired(s, *order)
	if err != nil {
		slog.Error("countRequired prev", slog.Any("error", err))
		return nil, err
	}
	if err := returnItems(s, requiredIngredientsPrev); err != nil {
		slog.Error("returnItems", slog.Any("error", err))
		return nil, err
	}
	requiredIngredientsNew, err := countRequired(s, updatedOrder)
	if err != nil {
		slog.Error("countRequired new", slog.Any("error", err))
		return nil, err
	}
	conflicts, err := compareIngredients(s, updatedOrder, requiredIngredientsNew)
	if err != nil {
		slog.Error("compareIngredients", slog.Any("error", err))
		_ = orderResult(s, requiredIngredientsPrev)
		return conflicts, err
	}
	if len(conflicts) != 0 {
		slog.Warn("update conflicts", slog.String("order_id", id), slog.Any("conflicts", conflicts))
		_ = orderResult(s, requiredIngredientsPrev)
		return conflicts, nil
	} else {
		if err := orderResult(s, requiredIngredientsNew); err != nil {
			slog.Error("orderResult", slog.Any("error", err))
			return nil, err
		}
	}
	if updatedOrder.ID == "" {
		updatedOrder.ID = id
	}
	if updatedOrder.CustomerName == "" {
		updatedOrder.CustomerName = order.CustomerName
	}
	if updatedOrder.CreatedAt == "" {
		updatedOrder.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if updatedOrder.Status == "" {
		updatedOrder.Status = order.Status
	}
	if len(updatedOrder.Items) == 0 {
		slog.Error("empty items", slog.String("order_id", id))
		return nil, fmt.Errorf("items cannot be empty")
	}
	if err := s.orderRepo.Update(id, updatedOrder); err != nil {
		slog.Error("Update order", slog.Any("error", err))
		return nil, err
	}
	slog.Info("Order updated", slog.String("order_id", id))
	return nil, nil
}

func returnItems(s *OrderServ, requiredIngredients map[string]float64) error {
	invItems, err := s.invRepo.FindAll()
	if err != nil {
		slog.Error("FindAll inventory", slog.Any("error", err))
		return err
	}
	for key, val := range requiredIngredients {
		for i := range invItems {
			if invItems[i].IngredientID == key {
				invItems[i].Quantity += val
				if err := s.invRepo.Update(key, invItems[i]); err != nil {
					slog.Error("Update inventory", slog.String("ingredient_id", key), slog.Any("error", err))
					return nil
				}
			}
		}
	}
	return nil
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
			}
			if _, err := s.invRepo.FindByID(ingredient.IngredientID); err == nil {
				requiredIngredients[ingredient.IngredientID] += ingredient.Quantity * float64(menuItem.Quantity)
			} else {
				requiredIngredients[ingredient.IngredientID] = -2
			}
		}
	}
	return requiredIngredients, nil
}

func compareIngredients(s *OrderServ, order models.Order, requiredIngredients map[string]float64) ([]string, error) {
	var conflicts []string
	for key, val := range requiredIngredients {
		if val == -1 {
			conflicts = append(conflicts, "ingredient "+key+" has negative value")
			continue
		}
		if val == -2 {
			conflicts = append(conflicts, "ingredient "+key+" not found")
			continue
		}
		invItem, err := s.invRepo.FindByID(key)
		if err != nil {
			return conflicts, err
		}
		if val > invItem.Quantity {
			conflicts = append(conflicts,
				"Insufficient inventory for ingredient '"+key+"'. Required: "+
					strconv.Itoa(int(val))+invItem.Unit+", Available: "+
					strconv.Itoa(int(invItem.Quantity))+invItem.Unit)
		}
	}
	return conflicts, nil
}

func orderResult(s *OrderServ, requiredIngredients map[string]float64) error {
	invItems, err := s.invRepo.FindAll()
	if err != nil {
		slog.Error("FindAll inventory", slog.Any("error", err))
		return err
	}
	for key, val := range requiredIngredients {
		for i := range invItems {
			if invItems[i].IngredientID == key {
				invItems[i].Quantity -= val
				if err := s.invRepo.Update(key, invItems[i]); err != nil {
					slog.Error("Update inventory", slog.String("ingredient_id", key), slog.Any("error", err))
					return nil
				}
			}
		}
	}
	return nil
}

func (s *OrderServ) DeleteOrder(id string) error {
	slog.Info("DeleteOrder", slog.String("order_id", id))
	err := s.orderRepo.Delete(id)
	if err != nil {
		slog.Error("DeleteOrder failed", slog.String("order_id", id), slog.Any("error", err))
		return err
	}
	return nil
}

func (s *OrderServ) CloseOrder(id string) error {
	slog.Info("CloseOrder", slog.String("order_id", id))
	err := s.orderRepo.Close(id)
	if err != nil {
		slog.Error("CloseOrder failed", slog.String("order_id", id), slog.Any("error", err))
		return err
	}
	return nil
}

func (s *OrderServ) GetTotalSales() (models.Total, error) {
	slog.Info("GetTotalSales")
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		slog.Error("FindAll orders", slog.Any("error", err))
		return models.Total{}, err
	}
	totalSales := 0.0
	for _, order := range orders {
		for _, menuItem := range order.Items {
			if order.Status == "closed" {
				item, err := s.menuRepo.FindByID(menuItem.ProductID)
				if err != nil {
					slog.Error("FindByID menu", slog.String("product_id", menuItem.ProductID), slog.Any("error", err))
					return models.Total{}, err
				}
				totalSales += item.Price * float64(menuItem.Quantity)
			}
		}
	}
	slog.Info("TotalSales", slog.Float64("total", totalSales))
	return models.Total{TotalSales: totalSales}, nil
}

func (s *OrderServ) GetPopularMenuItems() ([]string, error) {
	slog.Info("GetPopularMenuItems")
	orders, err := s.orderRepo.FindAll()
	if err != nil {
		slog.Error("FindAll orders", slog.Any("error", err))
		return nil, err
	}
	menuItems := make(map[string]int)
	for _, order := range orders {
		if order.Status == "closed" {
			for _, menuItem := range order.Items {
				menuItems[menuItem.ProductID] += menuItem.Quantity
			}
		}
	}
	keys := make([]string, 0, len(menuItems))
	for k := range menuItems {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return menuItems[keys[i]] > menuItems[keys[j]]
	})
	if len(keys) < 3 {
		err := fmt.Errorf("not enough menu items to achieve result")
		slog.Warn("not enough popular items", slog.Any("error", err))
		return nil, err
	}
	var popularMenuItems []string
	for i := 0; i < 3; i++ {
		row := keys[i] + ": " + strconv.Itoa(menuItems[keys[i]])
		popularMenuItems = append(popularMenuItems, row)
	}
	slog.Info("PopularMenuItems", slog.Any("items", popularMenuItems))
	return popularMenuItems, nil
}
