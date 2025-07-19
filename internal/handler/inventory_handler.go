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
	slog.Info("Inventory", slog.String("method", r.Method), slog.String("path", r.URL.Path))

	switch r.Method {
	case http.MethodGet:
		items, err := h.svc.GetAllInventoryItem()
		if err != nil {
			slog.Error("Inventory GET failed", slog.Any("error", err))
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		slog.Info("Inventory GET ok", slog.Int("count", len(items)))
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("Inventory GET encode failed", slog.Any("error", err))
		}

	case http.MethodPost:
		var item models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			slog.Warn("Inventory POST bad JSON", slog.Any("error", err))
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Info("Inventory POST decoded", slog.String("id", item.IngredientID))

		if item.IngredientID == "" {
			slog.Warn("Inventory POST validation", slog.String("field", "IngredientID"))
			writeJSONError(w, http.StatusBadRequest, "ingredient_id is required")
			return
		}
		if item.Name == "" {
			slog.Warn("Inventory POST validation", slog.String("field", "Name"))
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if item.Quantity < 0 {
			slog.Warn("Inventory POST validation", slog.Int("quantity", int(item.Quantity)))
			writeJSONError(w, http.StatusBadRequest, "quantity must be non‑negative")
			return
		}
		if item.Unit == "" {
			slog.Warn("Inventory POST validation", slog.String("field", "Unit"))
			writeJSONError(w, http.StatusBadRequest, "unit is required")
			return
		}

		if err := h.svc.AddInventoryItem(item); err != nil {
			slog.Warn("Inventory POST service error", slog.Any("error", err))
			switch err.Error() {
			case "Item ID already exists", "Item Name already exists":
				writeJSONError(w, http.StatusConflict, err.Error())
			default:
				writeJSONError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		slog.Info("Inventory POST created", slog.String("id", item.IngredientID))
		w.WriteHeader(http.StatusCreated)

	default:
		slog.Warn("Inventory unsupported method", slog.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func (h *InventoryHandler) InventoryByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/inventory/")
	slog.Info("InventoryByID", slog.String("method", r.Method), slog.String("id", id))

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetInventoryItemByID(id)
		if err != nil {
			slog.Warn("InventoryByID GET not found", slog.String("id", id))
			writeJSONError(w, http.StatusNotFound, "inventory item not found")
			return
		}
		slog.Info("InventoryByID GET ok", slog.String("id", id))
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(item); err != nil {
			slog.Error("InventoryByID GET encode failed", slog.Any("error", err))
		}

	case http.MethodPut:
		var updated models.InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			slog.Warn("InventoryByID PUT bad JSON", slog.Any("error", err))
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Info("InventoryByID PUT decoded", slog.String("id", id))

		if updated.IngredientID == "" {
			slog.Warn("InventoryByID PUT validation", slog.String("field", "IngredientID"))
			writeJSONError(w, http.StatusBadRequest, "ingredient_id is required")
			return
		}
		if updated.Name == "" {
			slog.Warn("InventoryByID PUT validation", slog.String("field", "Name"))
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if updated.Quantity < 0 {
			slog.Warn("InventoryByID PUT validation", slog.Int("quantity", int(updated.Quantity)))
			writeJSONError(w, http.StatusBadRequest, "quantity must be non‑negative")
			return
		}
		if updated.Unit == "" {
			slog.Warn("InventoryByID PUT validation", slog.String("field", "Unit"))
			writeJSONError(w, http.StatusBadRequest, "unit is required")
			return
		}

		if err := h.svc.UpdateInventoryItem(id, updated); err != nil {
			slog.Warn("InventoryByID PUT service error", slog.Any("error", err))
			switch err.Error() {
			case "Inventory item ID already exists", "Inventory item name already exists":
				writeJSONError(w, http.StatusConflict, err.Error())
			default:
				writeJSONError(w, http.StatusNotFound, err.Error())
			}
			return
		}

		slog.Info("InventoryByID PUT success", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.svc.DeleteInventoryItem(id); err != nil {
			slog.Warn("InventoryByID DELETE failed", slog.String("id", id), slog.Any("error", err))
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Info("InventoryByID DELETE success", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)

	default:
		slog.Warn("InventoryByID unsupported method", slog.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}
