package handler

import (
	"encoding/json"
	"hot-coffee/internal/service"
	"hot-coffee/models"
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
	switch r.Method {
	case http.MethodPost:
		var item models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.svc.AddMenuItem(item); err != nil {
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
		w.WriteHeader(http.StatusCreated)

	case http.MethodGet:
		items, err := h.svc.GetAllMenuItems()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(items)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *MenuHandler) MenuByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")

	switch r.Method {
	case http.MethodGet:
		item, err := h.svc.GetMenuItemByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var updatedItem models.MenuItem
		if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.svc.UpdateMenuItem(id, updatedItem); err != nil {
			switch err.Error() {
			case "Menu item ID already exists", "Menu item name already exists":
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusNotFound)
			}
			return
		}

	case http.MethodDelete:
		if err := h.svc.DeleteMenuItem(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
