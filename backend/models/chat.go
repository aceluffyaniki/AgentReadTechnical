package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatRequest is sent by client to /api/chat
type ChatRequest struct {
	SessionID  uuid.UUID `json:"session_id" binding:"required"`
	AnalysisID uuid.UUID `json:"analysis_id" binding:"required"`
	Message    string    `json:"message" binding:"required,min=1,max=4000"`
}

// ChatResponse is returned by /api/chat
type ChatResponse struct {
	SessionID  uuid.UUID `json:"session_id"`
	MessageID  uuid.UUID `json:"message_id"`
	Reply      string    `json:"reply"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	ID         uuid.UUID `db:"id"`
	SessionID  uuid.UUID `db:"session_id"`
	AnalysisID uuid.UUID `db:"analysis_id"`
	Role       string    `db:"role"`    // user | assistant
	Content    string    `db:"content"`
	TokenCount int       `db:"token_count"`
	CreatedAt  time.Time `db:"created_at"`
}
