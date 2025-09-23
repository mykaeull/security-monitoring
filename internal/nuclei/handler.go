package subdomainenum

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc Service) {
	mux.HandleFunc("/api/v1/subdomain-enum", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			subdomainEnumPostHandler(w, r, svc)
		case http.MethodGet:
			subdomainEnumGetHandler(w, r, svc)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func subdomainEnumPostHandler(w http.ResponseWriter, r *http.Request, svc Service) {
	var req Request
	if r.Body != nil {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&req)
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

func subdomainEnumGetHandler(w http.ResponseWriter, r *http.Request, svc Service) {
	records, err := svc.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(records)
}
