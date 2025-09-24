package nuclei

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc *NucleiService) {
	mux.HandleFunc("/api/v1/nuclei", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := svc.Run(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
