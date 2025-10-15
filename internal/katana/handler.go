package katana

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc *KatanaService) {
	mux.HandleFunc("/api/v1/katana", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Dispara o Katana
			if err := svc.Run(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodGet:
			// Opcional: se amanhã quiser listar algo (ex: últimas URLs crawled),
			// já fica preparado para expandir
			resp := map[string]string{"status": "katana service ready"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
