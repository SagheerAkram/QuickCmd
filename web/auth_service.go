package web

import (
	"fmt"
	"sync"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication and authorization
type AuthService struct {
	config *AuthConfig
	users  map[string]*User
	mu     sync.RWMutex
}

// NewAuthService creates a new auth service
func NewAuthService(config *AuthConfig) *AuthService {
	service := &AuthService{
		config: config,
		users:  make(map[string]*User),
	}
	
	// Add default dev users if in dev mode
	if config.DevMode {
		service.addDefaultUsers()
	}
	
	return service
}

// addDefaultUsers adds default users for development
func (s *AuthService) addDefaultUsers() {
	defaultUsers := []struct {
		username string
		password string
		roles    []Role
	}{
		{"admin", "admin", []Role{RoleAdmin}},
		{"approver", "approver", []Role{RoleApprover, RoleOperator}},
		{"operator", "operator", []Role{RoleOperator}},
		{"viewer", "viewer", []Role{RoleViewer}},
	}
	
	for _, u := range defaultUsers {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		user := &User{
			ID:       u.username,
			Username: u.username,
			Password: string(hashedPassword),
			Roles:    u.roles,
			Created:  time.Now(),
		}
		s.users[u.username] = user
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (string, error) {
	s.mu.RLock()
	user, exists := s.users[username]
	s.mu.RUnlock()
	
	if !exists {
		return "", ErrInvalidCredentials
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	
	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", err
	}
	
	return token, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "quickcmd",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	
	if err != nil {
		return nil, ErrInvalidToken
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check expiration
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, ErrTokenExpired
		}
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// GetUser retrieves a user by username
func (s *AuthService) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}
	
	return user, nil
}

// AddUser adds a new user (dev mode only)
func (s *AuthService) AddUser(username, password string, roles []Role) error {
	if !s.config.DevMode {
		return fmt.Errorf("adding users only allowed in dev mode")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.users[username]; exists {
		return fmt.Errorf("user already exists")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user := &User{
		ID:       username,
		Username: username,
		Password: string(hashedPassword),
		Roles:    roles,
		Created:  time.Now(),
	}
	
	s.users[username] = user
	return nil
}

// CheckRole verifies if a user has a specific role
func (s *AuthService) CheckRole(claims *Claims, requiredRole Role) error {
	for _, role := range claims.Roles {
		if role == requiredRole || role == RoleAdmin {
			return nil
		}
	}
	return ErrInsufficientRole
}

// CheckAnyRole verifies if a user has any of the specified roles
func (s *AuthService) CheckAnyRole(claims *Claims, requiredRoles ...Role) error {
	for _, required := range requiredRoles {
		for _, role := range claims.Roles {
			if role == required || role == RoleAdmin {
				return nil
			}
		}
	}
	return ErrInsufficientRole
}
