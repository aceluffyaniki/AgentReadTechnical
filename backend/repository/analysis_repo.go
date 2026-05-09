package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/agentrading/backend/models"
)

// SaveAnalysis persists a completed token analysis to the database.
func SaveAnalysis(ctx context.Context, req *models.AnalyzeRequest, scraper *models.ScraperResult, ai *models.AIAnalysis, usage *models.TokenUsage, modelUsed string) (*models.TokenAnalysisDB, error) {
	aiJSON, err := json.Marshal(ai)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI response: %w", err)
	}

	domJSON, err := json.Marshal(scraper.DomMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DOM metrics: %w", err)
	}

	id := uuid.New()
	now := time.Now()

	_, err = Pool.Exec(ctx, `
		INSERT INTO token_analysis (
			id, token_address, chain, chart_url, screenshot_b64,
			dom_metrics, ai_response,
			skor_scam, status_keamanan, smc_bias,
			model_used, prompt_tokens, output_tokens, cache_read_tokens,
			analyzed_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9, $10,
			$11, $12, $13, $14,
			$15
		)`,
		id, req.TokenAddress, req.Chain, req.ChartURL, scraper.ScreenshotB64,
		domJSON, aiJSON,
		ai.Keamanan.SkorScam, ai.Keamanan.Status, ai.SmartMoney.Bias,
		modelUsed, usage.PromptTokens, usage.OutputTokens, usage.CacheReadTokens,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert token_analysis: %w", err)
	}

	return &models.TokenAnalysisDB{
		ID:             id,
		TokenAddress:   req.TokenAddress,
		Chain:          req.Chain,
		SkorScam:       ai.Keamanan.SkorScam,
		StatusKeamanan: ai.Keamanan.Status,
		SMCBias:        ai.SmartMoney.Bias,
		ModelUsed:      modelUsed,
		AnalyzedAt:     now,
	}, nil
}

// GetLatestAnalysis fetches the most recent analysis for a given token+chain.
func GetLatestAnalysis(ctx context.Context, tokenAddress, chain string) (*models.TokenAnalysisDB, error) {
	row := Pool.QueryRow(ctx, `
		SELECT id, token_address, chain, chart_url, screenshot_b64,
		       dom_metrics, ai_response,
		       skor_scam, status_keamanan, smc_bias,
		       model_used, prompt_tokens, output_tokens, cache_read_tokens,
		       analyzed_at
		FROM token_analysis
		WHERE token_address = $1 AND chain = $2
		ORDER BY analyzed_at DESC
		LIMIT 1`,
		tokenAddress, chain,
	)

	var rec models.TokenAnalysisDB
	err := row.Scan(
		&rec.ID, &rec.TokenAddress, &rec.Chain, &rec.ChartURL, &rec.ScreenshotB64,
		&rec.DomMetrics, &rec.AIResponse,
		&rec.SkorScam, &rec.StatusKeamanan, &rec.SMCBias,
		&rec.ModelUsed, &rec.PromptTokens, &rec.OutputTokens, &rec.CacheReadTokens,
		&rec.AnalyzedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetLatestAnalysis: %w", err)
	}
	return &rec, nil
}

// GetAnalysisByID fetches a specific analysis record by its UUID.
func GetAnalysisByID(ctx context.Context, id uuid.UUID) (*models.TokenAnalysisDB, error) {
	row := Pool.QueryRow(ctx, `
		SELECT id, token_address, chain, chart_url, screenshot_b64,
		       dom_metrics, ai_response,
		       skor_scam, status_keamanan, smc_bias,
		       model_used, prompt_tokens, output_tokens, cache_read_tokens,
		       analyzed_at
		FROM token_analysis WHERE id = $1`,
		id,
	)

	var rec models.TokenAnalysisDB
	err := row.Scan(
		&rec.ID, &rec.TokenAddress, &rec.Chain, &rec.ChartURL, &rec.ScreenshotB64,
		&rec.DomMetrics, &rec.AIResponse,
		&rec.SkorScam, &rec.StatusKeamanan, &rec.SMCBias,
		&rec.ModelUsed, &rec.PromptTokens, &rec.OutputTokens, &rec.CacheReadTokens,
		&rec.AnalyzedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetAnalysisByID: %w", err)
	}
	return &rec, nil
}
