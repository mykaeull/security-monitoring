package httpxscan

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc *Service) {
	mux.HandleFunc("/api/v1/httpx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Executa o scan; não retorna corpo
		if err := svc.Run(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204
	})
}
