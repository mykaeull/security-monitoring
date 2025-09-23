package domain

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc Service) {
	mux.HandleFunc("/api/v1/domain", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			addDomainPostHandler(w, r, svc)
		case http.MethodGet:
			addDomainGetHandler(w, r, svc)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func addDomainPostHandler(w http.ResponseWriter, r *http.Request, svc Service) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
	}

	ins, skip, details, err := svc.AddDomains(r.Context(), req.Domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := Response{Inserted: ins, Skipped: skip, Details: details}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func addDomainGetHandler(w http.ResponseWriter, r *http.Request, svc Service) {
	records, err := svc.GetAllDomains(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(records)
}
