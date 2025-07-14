package handler

import (
	"encoding/json"
	"hot-coffee/internal/service"
	"hot-coffee/models"
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
	switch r.Method {
	case http.MethodGet:
		items, err := h.svc.GetAllInventoryItem()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(items)

	case http.MethodPost:
		var item models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.svc.AddInventoryItem(item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *InventoryHandler) InventoryByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/inventory/")

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetInventoryItemByID(id)
		if err != nil {
			http.Error(w, "inventory item not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var updatedItem models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.svc.UpdateInventoryItem(id, updatedItem); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

	case http.MethodDelete:
		if err := h.svc.DeleteInventoryItem(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
