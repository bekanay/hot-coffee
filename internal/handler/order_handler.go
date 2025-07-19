package handler

import (
	"encoding/json"
	"hot-coffee/internal/service"
	"hot-coffee/models"
	"net/http"
	"strings"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{svc: orderService}
}

func (h *OrderHandler) Orders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orders, err := h.svc.GetOrders()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	case http.MethodPost:
		var newOrder models.Order
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		conflicts, err := h.svc.CreateOrder(newOrder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(conflicts) != 0 {
			text := ""
			for _, conflict := range conflicts {
				text += conflict
				text += "\n"
			}
			http.Error(w, text, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func (h *OrderHandler) OrderByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 4 {
		if r.Method == http.MethodPost {
			if parts[3] != "close" {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			err := h.svc.CloseOrder(parts[2])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	} else if len(parts) == 3 {
		switch r.Method {
		case http.MethodGet:
			order, err := h.svc.GetOrderById(parts[2])
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "json/application")
			json.NewEncoder(w).Encode(order)
		case http.MethodPut:
			var order models.Order
			if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			conflicts, err := h.svc.UpdateOrder(parts[2], order)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				if err.Error() == "items cannot be empty" {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(conflicts) != 0 {
				text := ""
				for _, conflict := range conflicts {
					text += conflict
					text += "\n"
				}
				http.Error(w, text, http.StatusBadRequest)
				return
			}
		case http.MethodDelete:
			err := h.svc.DeleteOrder(parts[2])
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func (h *OrderHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	totalSales, err := h.svc.GetTotalSales()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totalSales)
}

func (h *OrderHandler) GetPopularMenuItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	popularMenuItems, err := h.svc.GetPopularMenuItems()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularMenuItems)
}
