package handler

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type AdminHandler struct {
	dataDir string
}

func NewAdminHandler(dir string) *AdminHandler {
	return &AdminHandler{dataDir: dir}
}

func (h *AdminHandler) ResetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	files := []string{
		filepath.Join(h.dataDir, "orders.json"),
		filepath.Join(h.dataDir, "menu_items.json"),
		filepath.Join(h.dataDir, "inventory.json"),
	}

	for _, path := range files {
		if err := ioutil.WriteFile(path, []byte("[]"), 0o644); err != nil {
			http.Error(w, "failed to clear "+filepath.Base(path)+": "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
