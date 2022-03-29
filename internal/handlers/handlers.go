package handlers

import (
	"net/http"
	"strings"
	"github.com/fedoroko/practicum_go/internal/storage"
)

func UpdateFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodPost{
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
        return
	}

	path := r.URL.String()
	pathArr := strings.Split(path, "/")
	if len(pathArr) != 5 {
		http.Error(w, "Wrong URL pattern. Should be: /update/<type>/<name>/<value>", http.StatusNotFound)
		return
	}

	if pathArr[2] != "gauge" && pathArr[2] != "counter" {
		http.Error(w, "Wrong type. Only types \"gauge\" and \"counter\" allowed", http.StatusNotImplemented)
		return
	}

	if pathArr[4] == "" {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}

	err := storage.Store(pathArr[2], pathArr[3], pathArr[4])
	if err != nil {
		http.Error(w, err.Error(), http.StatusOK)
	}

	w.WriteHeader(http.StatusOK)
}