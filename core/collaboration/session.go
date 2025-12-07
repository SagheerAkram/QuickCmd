package collaboration

import (
	"fmt"
	"sync"
	"time"
)

// CollaborationSession manages real-time collaboration
type CollaborationSession struct {
	ID           string
	Participants []*Participant
	Command      string
	Status       string
	mu           sync.RWMutex
}

// Participant represents a session participant
type Participant struct {
	UserID   string
	Role     string // "editor", "reviewer", "observer"
	JoinedAt time.Time
}

// NewCollaborationSession creates a new session
func NewCollaborationSession(id string) *CollaborationSession {
	return &CollaborationSession{
		ID:           id,
		Participants: []*Participant{},
		Status:       "active",
	}
}

// Join adds a participant to the session
func (cs *CollaborationSession) Join(userID, role string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.Participants = append(cs.Participants, &Participant{
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	})
}

// UpdateCommand updates the command being edited
func (cs *CollaborationSession) UpdateCommand(userID, command string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	// Check if user has edit permission
	if !cs.canEdit(userID) {
		return fmt.Errorf("user %s cannot edit", userID)
	}
	
	cs.Command = command
	return nil
}

func (cs *CollaborationSession) canEdit(userID string) bool {
	for _, p := range cs.Participants {
		if p.UserID == userID && (p.Role == "editor" || p.Role == "reviewer") {
			return true
		}
	}
	return false
}

// BroadcastUpdate broadcasts an update to all participants
func (cs *CollaborationSession) BroadcastUpdate(update string) {
	// Send update to all participants via WebSocket
}
