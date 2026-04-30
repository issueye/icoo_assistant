package memory

import "time"

const (
	TypeShortTerm      = "short_term"
	TypeLongTerm       = "long_term"
	TypeSessionSummary = "session_summary"
	TypeAIPersonality  = "ai_personality"
	TypeUserProfile    = "user_profile"
)

type Memory struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags,omitempty"`
	Importance  float64   `json:"importance"`
	SessionID   string    `json:"sessionId,omitempty"`
	AccessCount int       `json:"accessCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type StoreInput struct {
	ID         string
	Type       string
	Content    string
	Tags       []string
	Importance float64
	SessionID  string
}

type QueryInput struct {
	Type          string
	Query         string
	Tags          []string
	Limit         int
	MinImportance float64
	SessionID     string
}
