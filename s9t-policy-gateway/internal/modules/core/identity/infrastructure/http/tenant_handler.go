// internal/modules/core/identity/infrastructure/http/tenant_handler.go
package http

import (
	"encoding/json"
	"net/http"

	"s9t.os/internal/modules/core/identity/application/command"
	"s9t.os/internal/modules/core/identity/application/dto"
)

// TenantHandler manages HTTP requests for tenants
type TenantHandler struct {
	createTenantHandler *command.CreateTenantHandler
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(createHandler *command.CreateTenantHandler) *TenantHandler {
	return &TenantHandler{
		createTenantHandler: createHandler,
	}
}

// HandleCreate processes the POST request to create a new tenant
func (h *TenantHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if req.Name == "" || req.Slug == "" {
		http.Error(w, "name and slug are required", http.StatusBadRequest)
		return
	}

	// Execute application use case
	tenantDTO, err := h.createTenantHandler.Execute(r.Context(), req)
	if err != nil {
		// Log error internally, return generic message (or specific if mapped)
		http.Error(w, "failed to create tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenantDTO)
}
