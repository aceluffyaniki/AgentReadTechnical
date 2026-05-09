package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/agentrading/backend/models"
	"github.com/agentrading/backend/repository"
	"github.com/agentrading/backend/services/ai"
)

// ChatHandler handles POST /api/chat
// Flow: Load analysis context → Load chat history → Call Claude → Save both messages → Return reply
type ChatHandler struct {
	ai *ai.VertexClient
}

func NewChatHandler(aiClient *ai.VertexClient) *ChatHandler {
	return &ChatHandler{ai: aiClient}
}

func (h *ChatHandler) Handle(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	log.Printf("[CHAT] session=%s analysis=%s", req.SessionID, req.AnalysisID)

	// Step 1: Fetch the analysis record to use as context
	analysis, err := repository.GetAnalysisByID(ctx, req.AnalysisID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found: " + err.Error()})
		return
	}

	// Step 2: Build system context from the analysis
	systemContext, err := ai.BuildChatSystemContext(analysis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build context: " + err.Error()})
		return
	}

	// Step 3: Load chat history (last 30 messages to stay within context limit)
	history, err := repository.GetChatHistory(ctx, req.SessionID, 30)
	if err != nil {
		log.Printf("[CHAT] Could not load history: %v — starting fresh", err)
		history = []models.ChatMessage{} // Non-fatal
	}

	// Step 4: Save user message to DB
	_, err = repository.SaveChatMessage(ctx, req.SessionID, req.AnalysisID, "user", req.Message, 0)
	if err != nil {
		log.Printf("[CHAT] Failed to save user message: %v", err)
	}

	// Step 5: Send to Claude
	reply, outputTokens, err := h.ai.Chat(ctx, systemContext, history, req.Message)
	if err != nil {
		log.Printf("[CHAT] AI error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI chat failed: " + err.Error()})
		return
	}

	// Step 6: Save assistant reply to DB
	saved, err := repository.SaveChatMessage(ctx, req.SessionID, req.AnalysisID, "assistant", reply, outputTokens)
	if err != nil {
		log.Printf("[CHAT] Failed to save assistant message: %v", err)
	}

	messageID := uuid.New()
	if saved != nil {
		messageID = saved.ID
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		SessionID: req.SessionID,
		MessageID: messageID,
		Reply:     reply,
		CreatedAt: saved.CreatedAt,
	})
}

// GetChatHistory handles GET /api/chat/history?session_id=...
func (h *ChatHandler) GetHistory(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session_id"})
		return
	}

	history, err := repository.GetChatHistory(c.Request.Context(), sessionID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session_id": sessionID, "messages": history})
}
