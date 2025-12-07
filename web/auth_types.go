package web

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

// Role represents a user role
type Role string

const (
	RoleViewer   Role = "viewer"
	RoleOperator Role = "operator"
	RoleApprover Role = "approver"
	RoleAdmin    Role = "admin"
)

// User represents a user in the system
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"-"` // Never expose password
	Roles    []Role   `json:"roles"`
	Email    string   `json:"email,omitempty"`
	Created  time.Time `json:"created"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []Role   `json:"roles"`
	jwt.RegisteredClaims
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
	DevMode       bool
	OIDCEnabled   bool
	OIDCConfig    *OIDCConfig
}

// OIDCConfig represents OIDC configuration (placeholder for future)
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWTSecret:     generateSecret(),
		TokenDuration: 15 * time.Minute,
		DevMode:       true,
		OIDCEnabled:   false,
	}
}

// generateSecret generates a random secret
func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInsufficientRole   = errors.New("insufficient role")
	ErrUserNotFound       = errors.New("user not found")
)

// HasRole checks if user has a specific role
func (u *User) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if r == role || r == RoleAdmin {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (u *User) HasAnyRole(roles ...Role) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}
