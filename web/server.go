package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	
	"github.com/gorilla/mux"
	"github.com/yourusername/quickcmd/core/audit"
)

// Server represents the web API server
type Server struct {
	router         *mux.Router
	authService    *AuthService
	approvalStore  *ApprovalStore
	auditStore     *audit.SQLiteStore
	config         *Config
}

// Config represents server configuration
type Config struct {
	Port          int
	AuthConfig    *AuthConfig
	AuditDBPath   string
	ApprovalDBPath string
	CORSOrigins   []string
}

// NewServer creates a new web server
func NewServer(config *Config) (*Server, error) {
	// Create auth service
	authService := NewAuthService(config.AuthConfig)
	
	// Create approval store
	approvalStore, err := NewApprovalStore(config.ApprovalDBPath)
	if err != nil {
		return nil, err
	}
	
	// Open audit store
	auditStore, err := audit.NewSQLiteStore(config.AuditDBPath)
	if err != nil {
		return nil, err
	}
	
	server := &Server{
		router:        mux.NewRouter(),
		authService:   authService,
		approvalStore: approvalStore,
		auditStore:    auditStore,
		config:        config,
	}
	
	server.setupRoutes()
	
	return server, nil
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// Public endpoints
	api.HandleFunc("/login", s.handleLogin).Methods("POST")
	
	// Protected endpoints
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.authMiddleware)
	
	protected.HandleFunc("/history", s.handleHistory).Methods("GET")
	protected.HandleFunc("/run/{id}", s.handleRunDetail).Methods("GET")
	protected.HandleFunc("/run/{id}/replay", s.handleReplay).Methods("POST")
	
	// Approval endpoints (require approver role)
	protected.HandleFunc("/approvals", s.handleGetApprovals).Methods("GET")
	protected.Handle("/approvals/{id}/approve", s.requireRole(RoleApprover, http.HandlerFunc(s.handleApprove))).Methods("POST")
	protected.Handle("/approvals/{id}/reject", s.requireRole(RoleApprover, http.HandlerFunc(s.handleReject))).Methods("POST")
	
	// CORS middleware
	s.router.Use(s.corsMiddleware)
}

// Handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	token, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"expires_in": int(s.config.AuthConfig.TokenDuration.Seconds()),
	})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	
	filter := r.URL.Query().Get("filter")
	
	// Get history from audit store
	records, err := s.auditStore.GetHistory(limit, filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to fetch history")
		return
	}
	
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
		"count":   len(records),
	})
}

func (s *Server) handleRunDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid run ID")
		return
	}
	
	record, err := s.auditStore.GetRecordByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "Run not found")
		return
	}
	
	s.writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid run ID")
		return
	}
	
	// Get original record
	original, err := s.auditStore.GetRecordByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "Run not found")
		return
	}
	
	// Create new audit record for replay (dry-run)
	claims := r.Context().Value("claims").(*Claims)
	replayRecord := &audit.RunRecord{
		Timestamp:       audit.Now(),
		User:            claims.Username,
		Prompt:          original.Prompt + " (REPLAY DRY-RUN)",
		SelectedCommand: original.SelectedCommand,
		RiskLevel:       original.RiskLevel,
		Executed:        false, // Dry-run, not executed
	}
	
	if err := s.auditStore.LogExecution(replayRecord); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to create replay record")
		return
	}
	
	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Replay dry-run created",
		"dry_run": true,
	})
}

func (s *Server) handleGetApprovals(w http.ResponseWriter, r *http.Request) {
	approvals, err := s.approvalStore.GetPendingApprovals()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to fetch approvals")
		return
	}
	
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"approvals": approvals,
		"count":     len(approvals),
	})
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid approval ID")
		return
	}
	
	var req struct {
		Confirmation string `json:"confirmation"`
		Note         string `json:"note"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate confirmation
	expectedConfirmation := "APPROVE " + vars["id"]
	if req.Confirmation != expectedConfirmation {
		s.writeError(w, http.StatusBadRequest, "Invalid confirmation. Type: "+expectedConfirmation)
		return
	}
	
	claims := r.Context().Value("claims").(*Claims)
	
	if err := s.approvalStore.ApproveApproval(id, claims.Username, req.Confirmation, req.Note); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to approve")
		return
	}
	
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Approval granted",
		"approved_by": claims.Username,
	})
}

func (s *Server) handleReject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid approval ID")
		return
	}
	
	var req struct {
		Reason string `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Reason == "" {
		s.writeError(w, http.StatusBadRequest, "Rejection reason required")
		return
	}
	
	claims := r.Context().Value("claims").(*Claims)
	
	if err := s.approvalStore.RejectApproval(id, claims.Username, req.Reason); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to reject")
		return
	}
	
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Approval rejected",
		"rejected_by": claims.Username,
	})
}

// Helper methods

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteStatus(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]interface{}{
		"error": message,
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + strconv.Itoa(s.config.Port)
	return http.ListenAndServe(addr, s.router)
}
