package handlers

import (
	"net/http"
	"fmt"
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
	fmt.Println(len(pathArr))
	if len(pathArr) != 5 {
		http.Error(w, "Wrong URL pattern. Should be: /update/<type>/<name>/<value>", http.StatusBadRequest)
		return
	}

	if pathArr[2] != "gauge" && pathArr[2] != "counter" {
		http.Error(w, "Wrong type. Only types \"gauge\" and \"counter\" allowed", http.StatusBadRequest)
		return
	}

	if pathArr[4] == "" {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}

	err := storage.Store(pathArr[2], pathArr[3], pathArr[4])
	if err != nil {
		http.Error(w, err, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}