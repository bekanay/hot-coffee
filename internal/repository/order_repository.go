package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"sync"

	"hot-coffee/models"
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
		slog.Error("loadOrders: ReadFile failed", "path", path, "err", err)
		return nil, err
	}
	var orders []models.Order
	if err := json.Unmarshal(raw, &orders); err != nil {
		slog.Error("loadOrders: Unmarshal failed", "err", err)
		return nil, err
	}
	slog.Info("loadOrders: success", "count", len(orders))
	return orders, nil
}

func (r *jsonOrderRepo) saveOrders(orders []models.Order) error {
	path := filepath.Join(r.dataDir, "orders.json")
	raw, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		slog.Error("saveOrders: MarshalIndent failed", "err", err)
		return err
	}
	if err := ioutil.WriteFile(path, raw, 0o644); err != nil {
		slog.Error("saveOrders: WriteFile failed", "path", path, "err", err)
		return err
	}
	slog.Info("saveOrders: success", "count", len(orders))
	return nil
}

func (r *jsonOrderRepo) Add(order models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Add order", "orderID", order.ID)
	orders, err := r.loadOrders()
	if err != nil {
		return err
	}
	orders = append(orders, order)
	if err := r.saveOrders(orders); err != nil {
		slog.Error("Add: saveOrders failed", "err", err)
		return err
	}
	slog.Info("Add: success", "orderID", order.ID)
	return nil
}

func (r *jsonOrderRepo) FindAll() ([]models.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders, err := r.loadOrders()
	if err != nil {
		slog.Error("FindAll: loadOrders failed", "err", err)
		return nil, err
	}
	slog.Info("FindAll: returning orders", "count", len(orders))
	return orders, nil
}

func (r *jsonOrderRepo) FindByID(id string) (*models.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("FindByID", "orderID", id)
	orders, err := r.loadOrders()
	if err != nil {
		slog.Error("FindByID: loadOrders failed", "err", err)
		return nil, err
	}
	for _, o := range orders {
		if o.ID == id {
			slog.Info("FindByID: found", "orderID", id)
			return &o, nil
		}
	}
	slog.Warn("FindByID: not found", "orderID", id)
	return nil, fmt.Errorf("order %s not found", id)
}

func (r *jsonOrderRepo) Update(id string, updated models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Update order", "orderID", id)
	orders, err := r.loadOrders()
	if err != nil {
		return err
	}
	for i, o := range orders {
		if o.ID == id {
			orders[i] = updated
			if err := r.saveOrders(orders); err != nil {
				slog.Error("Update: saveOrders failed", "err", err)
				return err
			}
			slog.Info("Update: success", "orderID", id)
			return nil
		}
	}
	slog.Warn("Update: not found", "orderID", id)
	return fmt.Errorf("order %s not found", id)
}

func (r *jsonOrderRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Delete order", "orderID", id)
	orders, err := r.loadOrders()
	if err != nil {
		return err
	}
	filtered := orders[:0]
	for _, o := range orders {
		if o.ID != id {
			filtered = append(filtered, o)
		}
	}
	if len(filtered) == len(orders) {
		slog.Warn("Delete: not found", "orderID", id)
		return fmt.Errorf("order %s not found", id)
	}
	if err := r.saveOrders(filtered); err != nil {
		slog.Error("Delete: saveOrders failed", "err", err)
		return err
	}
	slog.Info("Delete: success", "orderID", id)
	return nil
}

func (r *jsonOrderRepo) Close(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Close order", "orderID", id)
	orders, err := r.loadOrders()
	if err != nil {
		return err
	}
	for i, o := range orders {
		if o.ID == id {
			orders[i].Status = "closed"
			if err := r.saveOrders(orders); err != nil {
				slog.Error("Close: saveOrders failed", "err", err)
				return err
			}
			slog.Info("Close: success", "orderID", id)
			return nil
		}
	}
	slog.Warn("Close: not found", "orderID", id)
	return fmt.Errorf("order %s not found", id)
}
