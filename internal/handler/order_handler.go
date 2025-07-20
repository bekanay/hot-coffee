package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"hot-coffee/internal/service"
	"hot-coffee/models"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{svc: orderService}
}

func (h *OrderHandler) Orders(w http.ResponseWriter, r *http.Request) {
	slog.Info("Orders endpoint hit",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	switch r.Method {
	case http.MethodGet:
		orders, err := h.svc.GetOrders()
		if err != nil {
			slog.Error("GetOrders failed",
				slog.Any("error", err),
			)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			slog.Error("Encode orders failed",
				slog.Any("error", err),
			)
		}

	case http.MethodPost:
		var newOrder models.Order
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			slog.Error("Decode new order failed",
				slog.Any("error", err),
			)
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		conflicts, err := h.svc.CreateOrder(newOrder)
		if err != nil {
			slog.Error("CreateOrder failed",
				slog.Any("error", err),
			)
			if strings.Contains(err.Error(), "invalid quantity") {
				writeJSONError(w, http.StatusBadRequest, err.Error())
			} else {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		if len(conflicts) != 0 {
			msg := strings.Join(conflicts, "\n")
			slog.Warn("CreateOrder conflicts",
				slog.String("conflicts", msg),
			)
			writeJSONError(w, http.StatusBadRequest, msg)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func (h *OrderHandler) OrderByID(w http.ResponseWriter, r *http.Request) {
	slog.Info("OrderByID endpoint hit",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	parts := strings.Split(r.URL.Path, "/")
	// POST /orders/{id}/close
	if len(parts) == 4 && r.Method == http.MethodPost && parts[3] == "close" {
		if err := h.svc.CloseOrder(parts[2]); err != nil {
			slog.Error("CloseOrder failed",
				slog.String("order_id", parts[2]),
				slog.Any("error", err),
			)
			writeJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// GET, PUT, DELETE /orders/{id}
	if len(parts) == 3 {
		id := parts[2]
		switch r.Method {
		case http.MethodGet:
			order, err := h.svc.GetOrderById(id)
			if err != nil {
				slog.Error("GetOrderById failed",
					slog.String("order_id", id),
					slog.Any("error", err),
				)
				if strings.Contains(err.Error(), "not found") {
					writeJSONError(w, http.StatusNotFound, err.Error())
				} else {
					writeJSONError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(order); err != nil {
				slog.Error("Encode order failed",
					slog.Any("error", err),
				)
			}

		case http.MethodPut:
			var upd models.Order
			if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
				slog.Error("Decode update order failed",
					slog.Any("error", err),
				)
				writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}
			conflicts, err := h.svc.UpdateOrder(id, upd)
			if err != nil {
				slog.Error("UpdateOrder failed",
					slog.String("order_id", id),
					slog.Any("error", err),
				)
				switch {
				case strings.Contains(err.Error(), "not found"):
					writeJSONError(w, http.StatusNotFound, err.Error())
				case err.Error() == "items cannot be empty":
					writeJSONError(w, http.StatusBadRequest, err.Error())
				default:
					writeJSONError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
			if len(conflicts) != 0 {
				msg := strings.Join(conflicts, "\n")
				slog.Warn("UpdateOrder conflicts",
					slog.String("order_id", id),
					slog.String("conflicts", msg),
				)
				writeJSONError(w, http.StatusBadRequest, msg)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := h.svc.DeleteOrder(id); err != nil {
				slog.Error("DeleteOrder failed",
					slog.String("order_id", id),
					slog.Any("error", err),
				)
				if strings.Contains(err.Error(), "not found") {
					writeJSONError(w, http.StatusNotFound, err.Error())
				} else {
					writeJSONError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
		return
	}

	writeJSONError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (h *OrderHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	slog.Info("GetTotalSales endpoint hit",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	totalSales, err := h.svc.GetTotalSales()
	if err != nil {
		slog.Error("GetTotalSales failed",
			slog.Any("error", err),
		)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(totalSales); err != nil {
		slog.Error("Encode totalSales failed",
			slog.Any("error", err),
		)
	}
}

func (h *OrderHandler) GetPopularMenuItems(w http.ResponseWriter, r *http.Request) {
	slog.Info("GetPopularMenuItems endpoint hit",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
	items, err := h.svc.GetPopularMenuItems()
	if err != nil {
		slog.Error("GetPopularMenuItems failed",
			slog.Any("error", err),
		)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error("Encode popular items failed",
			slog.Any("error", err),
		)
	}
}
