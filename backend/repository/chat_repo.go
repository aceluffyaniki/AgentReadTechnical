package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/agentrading/backend/models"
)

// SaveChatMessage persists a single chat message to the database.
func SaveChatMessage(ctx context.Context, sessionID, analysisID uuid.UUID, role, content string, tokenCount int) (*models.ChatMessage, error) {
	id := uuid.New()
	now := time.Now()

	_, err := Pool.Exec(ctx, `
		INSERT INTO chat_history (id, session_id, analysis_id, role, content, token_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, sessionID, analysisID, role, content, tokenCount, now,
	)
	if err != nil {
		return nil, fmt.Errorf("SaveChatMessage: %w", err)
	}

	return &models.ChatMessage{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// GetChatHistory fetches ordered chat history for a session (oldest first).
// Limits to last N messages to avoid context overflow.
func GetChatHistory(ctx context.Context, sessionID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 30
	}

	rows, err := Pool.Query(ctx, `
		SELECT id, session_id, analysis_id, role, content, token_count, created_at
		FROM chat_history
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("GetChatHistory query: %w", err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.AnalysisID,
			&msg.Role, &msg.Content, &msg.TokenCount, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetChatHistory scan: %w", err)
		}
		messages = append(messages, msg)
	}

	// Reverse so oldest messages are first (chronological order)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteChatSession removes all messages for a session (for reset/clear feature).
func DeleteChatSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := Pool.Exec(ctx, `DELETE FROM chat_history WHERE session_id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("DeleteChatSession: %w", err)
	}
	return nil
}
