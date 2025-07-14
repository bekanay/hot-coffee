package handler

import (
	"encoding/json"
	"hot-coffee/internal/service"
	"hot-coffee/models"
	"log/slog"
	"net/http"
	"strings"
)

type InventoryHandler struct {
	svc service.InventoryService
}

func NewInventoryHandler(invService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: invService}
}

func (h *InventoryHandler) Inventory(w http.ResponseWriter, r *http.Request) {
	slog.Info("Inventory endpoint called", "method", r.Method, "path", r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		items, err := h.svc.GetAllInventoryItem()
		if err != nil {
			slog.Error("Inventory GET failed", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("Inventory GET success", "count", len(items))
		json.NewEncoder(w).Encode(items)

	case http.MethodPost:
		var item models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			slog.Warn("Inventory POST bad request", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("Inventory POST decoded", "id", item.IngredientID)
		if err := h.svc.AddInventoryItem(item); err != nil {
			slog.Warn("Inventory POST service error", "err", err)
			switch err.Error() {
			case "Item ID already exists", "Item Name already exists":
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		slog.Info("Inventory POST success", "id", item.IngredientID)
		w.WriteHeader(http.StatusCreated)

	default:
		slog.Warn("Inventory unsupported method", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *InventoryHandler) InventoryByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/inventory/")
	slog.Info("InventoryByID endpoint called", "method", r.Method, "id", id)

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetInventoryItemByID(id)
		if err != nil {
			slog.Warn("InventoryByID GET not found", "id", id)
			http.Error(w, "inventory item not found", http.StatusNotFound)
			return
		}
		slog.Info("InventoryByID GET success", "id", id)
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var updatedItem models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
			slog.Warn("InventoryByID PUT bad request", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("InventoryByID PUT decoded", "id", id)
		if err := h.svc.UpdateInventoryItem(id, updatedItem); err != nil {
			slog.Warn("InventoryByID PUT service error", "err", err)
			switch err.Error() {
			case "Inventory item ID already exists", "Inventory item name already exists":
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusNotFound)
			}
			return
		}
		slog.Info("InventoryByID PUT success", "id", id)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.svc.DeleteInventoryItem(id); err != nil {
			slog.Warn("InventoryByID DELETE not found", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Info("InventoryByID DELETE success", "id", id)
		w.WriteHeader(http.StatusNoContent)

	default:
		slog.Warn("InventoryByID unsupported method", "method", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
