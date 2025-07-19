package handler

import (
	"encoding/json"
	"hot-coffee/internal/service"
	"hot-coffee/models"
	"log/slog"
	"net/http"
	"strings"
)

type MenuHandler struct {
	svc service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{svc: menuService}
}

func (h *MenuHandler) Menu(w http.ResponseWriter, r *http.Request) {
	slog.Info("Menu", slog.String("method", r.Method), slog.String("path", r.URL.Path))
	switch r.Method {
	case http.MethodPost:
		var item models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			slog.Warn("Menu POST decode", slog.Any("error", err))
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Info("Menu POST decoded", slog.String("id", item.ID))

		if item.ID == "" {
			slog.Warn("Menu POST validation", slog.String("field", "ID"))
			writeJSONError(w, http.StatusBadRequest, "product_id is required")
			return
		}
		if item.Name == "" {
			slog.Warn("Menu POST validation", slog.String("field", "Name"))
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if item.Description == "" {
			slog.Warn("Menu POST validation", slog.String("field", "Description"))
			writeJSONError(w, http.StatusBadRequest, "description is required")
			return
		}
		if item.Price < 0 {
			slog.Warn("Menu POST validation", slog.Float64("price", item.Price))
			writeJSONError(w, http.StatusBadRequest, "price must be non-negative")
			return
		}
		if len(item.Ingredients) == 0 {
			slog.Warn("Menu POST validation", slog.String("field", "Ingredients"))
			writeJSONError(w, http.StatusBadRequest, "ingredients cannot be empty")
			return
		}
		for _, ing := range item.Ingredients {
			if ing.IngredientID == "" {
				slog.Warn("Menu POST validation", slog.String("field", "IngredientID"))
				writeJSONError(w, http.StatusBadRequest, "ingredient_id is required")
				return
			}
			if ing.Quantity <= 0 {
				slog.Warn("Menu POST validation", slog.String("ingredient_id", ing.IngredientID), slog.Int("quantity", int(ing.Quantity)))
				writeJSONError(w, http.StatusBadRequest, "ingredient quantity must be positive")
				return
			}
		}

		if err := h.svc.AddMenuItem(item); err != nil {
			slog.Warn("Menu POST service", slog.Any("error", err))
			switch {
			case strings.Contains(err.Error(), "following ingredients missing:"):
				writeJSONError(w, http.StatusNotFound, err.Error())
			case err.Error() == "Menu item ID already exists", err.Error() == "Menu item name already exists":
				writeJSONError(w, http.StatusConflict, err.Error())
			case err.Error() == "quantity must be non-negative value":
				writeJSONError(w, http.StatusBadRequest, err.Error())
			default:
				writeJSONError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		slog.Info("Menu POST success", slog.String("id", item.ID))
		w.WriteHeader(http.StatusCreated)

	case http.MethodGet:
		slog.Info("Menu GET", slog.String("method", r.Method))
		items, err := h.svc.GetAllMenuItems()
		if err != nil {
			slog.Error("Menu GET failed", slog.Any("error", err))
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		slog.Info("Menu GET success", slog.Int("count", len(items)))
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(items); err != nil {
			slog.Error("Menu GET encode", slog.Any("error", err))
		}

	default:
		slog.Warn("Menu unsupported", slog.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func (h *MenuHandler) MenuByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")
	slog.Info("MenuByID", slog.String("method", r.Method), slog.String("id", id))
	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetMenuItemByID(id)
		if err != nil {
			slog.Warn("MenuByID GET not found", slog.String("id", id))
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Info("MenuByID GET success", slog.String("id", id))
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(item); err != nil {
			slog.Error("MenuByID GET encode", slog.Any("error", err))
		}

	case http.MethodPut:
		var updated models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			slog.Warn("MenuByID PUT decode", slog.Any("error", err))
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Info("MenuByID PUT decoded", slog.String("id", id))

		if updated.ID == "" {
			slog.Warn("MenuByID PUT validation", slog.String("field", "ID"))
			writeJSONError(w, http.StatusBadRequest, "product_id is required")
			return
		}
		if updated.Name == "" {
			slog.Warn("MenuByID PUT validation", slog.String("field", "Name"))
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if updated.Description == "" {
			slog.Warn("MenuByID PUT validation", slog.String("field", "Description"))
			writeJSONError(w, http.StatusBadRequest, "description is required")
			return
		}
		if updated.Price < 0 {
			slog.Warn("MenuByID PUT validation", slog.Float64("price", updated.Price))
			writeJSONError(w, http.StatusBadRequest, "price must be non-negative")
			return
		}
		if len(updated.Ingredients) == 0 {
			slog.Warn("MenuByID PUT validation", slog.String("field", "Ingredients"))
			writeJSONError(w, http.StatusBadRequest, "ingredients cannot be empty")
			return
		}
		for _, ing := range updated.Ingredients {
			if ing.IngredientID == "" {
				slog.Warn("MenuByID PUT validation", slog.String("field", "IngredientID"))
				writeJSONError(w, http.StatusBadRequest, "ingredient_id is required")
				return
			}
			if ing.Quantity <= 0 {
				slog.Warn("MenuByID PUT validation", slog.String("ingredient_id", ing.IngredientID), slog.Int("quantity", int(ing.Quantity)))
				writeJSONError(w, http.StatusBadRequest, "ingredient quantity must be positive")
				return
			}
		}

		if err := h.svc.UpdateMenuItem(id, updated); err != nil {
			slog.Warn("MenuByID PUT service", slog.Any("error", err))
			switch {
			case err.Error() == "Menu item ID already exists", err.Error() == "Menu item name already exists":
				writeJSONError(w, http.StatusConflict, err.Error())
			default:
				writeJSONError(w, http.StatusNotFound, err.Error())
			}
			return
		}

		slog.Info("MenuByID PUT success", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.svc.DeleteMenuItem(id); err != nil {
			slog.Warn("MenuByID DELETE failed", slog.String("id", id), slog.Any("error", err))
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Info("MenuByID DELETE success", slog.String("id", id))
		w.WriteHeader(http.StatusNoContent)

	default:
		slog.Warn("MenuByID unsupported", slog.String("method", r.Method))
		writeJSONError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}
