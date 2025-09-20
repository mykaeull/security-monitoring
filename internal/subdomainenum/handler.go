package subdomainenum

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc Service) {
	// Injeta o service via closure
	mux.HandleFunc("/api/v1/subdomain-enum", func(w http.ResponseWriter, r *http.Request) {
		subdomainEnumHandler(w, r, svc)
	})
}

func subdomainEnumHandler(w http.ResponseWriter, r *http.Request, svc Service) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if r.Body != nil {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&req) // body opcional
	}

	status, body, contentType, err := svc.Enumerate(r.Context(), req.ApexDomain)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
