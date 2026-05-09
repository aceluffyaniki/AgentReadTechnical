package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/agentrading/backend/models"
	"github.com/agentrading/backend/repository"
	"github.com/agentrading/backend/services"
	"github.com/agentrading/backend/services/ai"
)

// AnalyzeHandler handles POST /api/analyze
// Flow: Scrape → AI analyze → Save to DB → Return result
type AnalyzeHandler struct {
	scraper *services.ScraperClient
	ai      *ai.VertexClient
}

func NewAnalyzeHandler(scraper *services.ScraperClient, aiClient *ai.VertexClient) *AnalyzeHandler {
	return &AnalyzeHandler{scraper: scraper, ai: aiClient}
}

func (h *AnalyzeHandler) Handle(c *gin.Context) {
	var req models.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	log.Printf("[ANALYZE] Start: token=%s chain=%s", req.TokenAddress, req.Chain)

	// Step 1: Call Python scraper microservice
	scraperResult, err := h.scraper.Scrape(ctx, &req)
	if err != nil {
		log.Printf("[ANALYZE] Scraper error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scraping failed: " + err.Error()})
		return
	}
	log.Printf("[ANALYZE] Scraper done — source_type:%s screenshot:%v",
		scraperResult.SourceType, scraperResult.ScreenshotB64 != "")

	// Step 2: Build text prompt
	userPrompt := ai.BuildAnalyzeUserPrompt(
		scraperResult.DomMetrics,
		scraperResult.SourceCode,
		scraperResult.SourceType,
		req.Chain,
	)

	// Step 3: Send multimodal request to Claude via Vertex AI
	analysis, usage, err := h.ai.AnalyzeToken(ctx, scraperResult.ScreenshotB64, userPrompt)
	if err != nil {
		log.Printf("[ANALYZE] AI error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	// Step 4: Persist to PostgreSQL (skip if DB not available)
	if repository.Pool == nil {
		log.Printf("[ANALYZE] DB not available — skipping history save")
		c.JSON(http.StatusOK, models.AnalyzeResponse{
			TokenAddress: req.TokenAddress,
			Chain:        req.Chain,
			Result:       analysis,
			TokenUsage:   usage,
		})
		return
	}

	dbRecord, err := repository.SaveAnalysis(ctx, &req, scraperResult, analysis, usage, "claude-opus-4-5")
	if err != nil {
		log.Printf("[ANALYZE] DB save error: %v", err)
		// Non-fatal: return result even if DB fails
		c.JSON(http.StatusOK, models.AnalyzeResponse{
			TokenAddress: req.TokenAddress,
			Chain:        req.Chain,
			Result:       analysis,
			TokenUsage:   usage,
		})
		return
	}

	log.Printf("[ANALYZE] Complete — id:%s skor_scam:%d status:%s",
		dbRecord.ID, dbRecord.SkorScam, dbRecord.StatusKeamanan)

	c.JSON(http.StatusOK, models.AnalyzeResponse{
		AnalysisID:   dbRecord.ID,
		TokenAddress: req.TokenAddress,
		Chain:        req.Chain,
		Result:       analysis,
		AnalyzedAt:   dbRecord.AnalyzedAt,
		TokenUsage:   usage,
	})
}
