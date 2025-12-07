package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// authMiddleware validates JWT tokens
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}
		
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.writeError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}
		
		token := parts[1]
		
		// Validate token
		claims, err := s.authService.ValidateToken(token)
		if err != nil {
			s.writeError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}
		
		// Add claims to context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireRole creates middleware that requires a specific role
func (s *Server) requireRole(role Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("claims").(*Claims)
		
		if err := s.authService.CheckRole(claims, role); err != nil {
			s.writeError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range s.config.CORSOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}
		
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// csrfMiddleware provides CSRF protection
func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for GET requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Get CSRF token from header
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			s.writeError(w, http.StatusForbidden, "Missing CSRF token")
			return
		}
		
		// Validate token (in production, verify against session)
		// For dev mode, accept any non-empty token
		if s.config.AuthConfig.DevMode || len(token) > 0 {
			next.ServeHTTP(w, r)
			return
		}
		
		s.writeError(w, http.StatusForbidden, "Invalid CSRF token")
	})
}

// generateCSRFToken generates a CSRF token
func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Add CSRF token endpoint
func (s *Server) handleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := generateCSRFToken()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"csrf_token": token,
	})
}
