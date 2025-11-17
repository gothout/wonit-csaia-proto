package queue

import (
	"encoding/json"
	"net/http"
	"strings"
)

// NewJobStatusHandler retorna um handler simples para consulta de status por ID.
func NewJobStatusHandler(mgr *JobManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		messageID := strings.TrimPrefix(r.URL.Path, "/jobs/")
		if messageID == "" {
			http.Error(w, "messageId requerido", http.StatusBadRequest)
			return
		}

		info, ok := mgr.Get(messageID)
		if !ok {
			http.Error(w, "job nao encontrado", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	})
}
