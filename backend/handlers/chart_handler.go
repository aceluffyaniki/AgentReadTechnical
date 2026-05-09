package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/agentrading/backend/services/ai"
)

// ChartAnalyzeHandler handles POST /api/analyze-chart
type ChartAnalyzeHandler struct {
	ai *ai.VertexClient
}

func NewChartAnalyzeHandler(aiClient *ai.VertexClient) *ChartAnalyzeHandler {
	return &ChartAnalyzeHandler{ai: aiClient}
}

type ChartAnalyzeRequest struct {
	ImageB64  string `json:"image_b64"`   // base64 screenshot (optional if chart_url set)
	ChartURL  string `json:"chart_url"`   // URL to screenshot (optional if image_b64 set)
	Notes     string `json:"notes"`       // user notes/context (optional)
	Timeframe string `json:"timeframe"`   // e.g. "1h", "4h", "1d"
	Source    string `json:"source"`      // e.g. "TradingView", "Binance", "Dexscreener"
}

type ChartAnalyzeResponse struct {
	Result     interface{} `json:"result"`
	TokenUsage interface{} `json:"token_usage,omitempty"`
}

func (h *ChartAnalyzeHandler) Handle(c *gin.Context) {
	var req ChartAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ImageB64 == "" && req.ChartURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Harus menyertakan image_b64 atau chart_url"})
		return
	}

	ctx := c.Request.Context()
	log.Printf("[CHART] Start — source:%s timeframe:%s hasImage:%v hasURL:%v",
		req.Source, req.Timeframe, req.ImageB64 != "", req.ChartURL != "")

	userPrompt := ai.BuildChartUserPrompt(req.Notes, req.Timeframe, req.Source)

	result, usage, err := h.ai.AnalyzeChart(ctx, req.ImageB64, req.ChartURL, userPrompt)
	if err != nil {
		log.Printf("[CHART] AI error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chart analysis failed: " + err.Error()})
		return
	}

	log.Printf("[CHART] Done — aksi:%v confidence:%v", getField(result, "aksi"), getField(result, "confidence"))

	c.JSON(http.StatusOK, ChartAnalyzeResponse{
		Result:     result,
		TokenUsage: usage,
	})
}

// getField safely reads a key from a map[string]interface{}
func getField(m interface{}, key string) interface{} {
	if mm, ok := m.(map[string]interface{}); ok {
		return mm[key]
	}
	return nil
}
