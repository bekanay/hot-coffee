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
	slog.Info("Menu endpoint called", "method", r.Method, "path", r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		var item models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			slog.Warn("Menu POST bad request", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("Menu POST decoded", "id", item.ID)

		if err := h.svc.AddMenuItem(item); err != nil {
			slog.Warn("Menu POST service error", "err", err)
			if strings.Contains(err.Error(), "following ingredients missing:") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			switch err.Error() {
			case "Menu item ID already exists", "Menu item name already exists":
				http.Error(w, err.Error(), http.StatusConflict)
			case "quantity must be non-negative value":
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		slog.Info("Menu POST success", "id", item.ID)
		w.WriteHeader(http.StatusCreated)

	case http.MethodGet:
		slog.Info("Menu GET called")
		items, err := h.svc.GetAllMenuItems()
		if err != nil {
			slog.Error("Menu GET failed", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("Menu GET success", "count", len(items))
		json.NewEncoder(w).Encode(items)

	default:
		slog.Warn("Menu unsupported method", "method", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *MenuHandler) MenuByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")
	slog.Info("MenuByID endpoint called", "method", r.Method, "id", id)

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetMenuItemByID(id)
		if err != nil {
			slog.Warn("MenuByID GET not found", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Info("MenuByID GET success", "id", id)
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var updatedItem models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
			slog.Warn("MenuByID PUT bad request", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("MenuByID PUT decoded", "id", id)

		if err := h.svc.UpdateMenuItem(id, updatedItem); err != nil {
			slog.Warn("MenuByID PUT service error", "err", err)
			switch err.Error() {
			case "Menu item ID already exists", "Menu item name already exists":
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusNotFound)
			}
			return
		}
		slog.Info("MenuByID PUT success", "id", id)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.svc.DeleteMenuItem(id); err != nil {
			slog.Warn("MenuByID DELETE not found", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Info("MenuByID DELETE success", "id", id)
		w.WriteHeader(http.StatusNoContent)

	default:
		slog.Warn("MenuByID unsupported method", "method", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
