package repository

import (
	"encoding/json"
	"fmt"
	"hot-coffee/models"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type OrderRepository interface {
	Add(order models.Order) error
	FindAll() ([]models.Order, error)
	FindByID(id string) (*models.Order, error)
	Update(id string, updated models.Order) error
	Delete(id string) error
	Close(id string) error
}

type jsonOrderRepo struct {
	dataDir string
	mu      sync.Mutex
}

func NewJSONOrderRepo(dir string) OrderRepository {
	return &jsonOrderRepo{dataDir: dir}
}

func (r *jsonOrderRepo) loadOrders() ([]models.Order, error) {
	path := filepath.Join(r.dataDir, "orders.json")

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := json.Unmarshal(raw, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *jsonOrderRepo) saveOrders(orders []models.Order) error {
	raw, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(r.dataDir, "orders.json")
	return ioutil.WriteFile(path, raw, 0644)
}

func (r *jsonOrderRepo) Add(order models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return err
	}

	orders = append(orders, order)
	return r.saveOrders(orders)
}

func (r *jsonOrderRepo) FindAll() ([]models.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.loadOrders()
}

func (r *jsonOrderRepo) FindByID(id string) (*models.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		if order.ID == id {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("order %s not found", id)
}

func (r *jsonOrderRepo) Update(id string, updated models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return err
	}

	for i, order := range orders {
		if order.ID == id {
			orders[i] = updated
			return r.saveOrders(orders)
		}
	}

	return fmt.Errorf("order %s not found", id)
}

func (r *jsonOrderRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return err
	}

	filtered := orders[:0]
	for _, order := range orders {
		if order.ID != id {
			filtered = append(filtered, order)
		}
	}

	if len(filtered) == len(orders) {
		return fmt.Errorf("order %s not found", id)
	}

	return r.saveOrders(filtered)
}

func (r *jsonOrderRepo) Close(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		return err
	}

	for i, order := range orders {
		if order.ID == id {
			orders[i].Status = "closed"
			return r.saveOrders(orders)
		}
	}

	return fmt.Errorf("order %s not found", id)
}
