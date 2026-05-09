package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/agentrading/backend/config"
	"github.com/agentrading/backend/models"
)

type ScraperClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewScraperClient() *ScraperClient {
	return &ScraperClient{
		baseURL: config.AppConfig.ScraperURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type ScrapeRequest struct {
	Chain        string `json:"chain"`
	TokenAddress string `json:"token_address"`
	ChartURL     string `json:"chart_url"`
}

// Scrape calls the Python Playwright microservice.
// Tolerant of partial errors — if scraper returned some useful data
// (screenshot or metrics), we proceed instead of failing hard.
func (s *ScraperClient) Scrape(ctx context.Context, req *models.AnalyzeRequest) (*models.ScraperResult, error) {
	body, err := json.Marshal(ScrapeRequest{
		Chain:        req.Chain,
		TokenAddress: req.TokenAddress,
		ChartURL:     req.ChartURL,
	})
	if err != nil {
		return nil, fmt.Errorf("ScraperClient: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/scrape", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ScraperClient: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ScraperClient: HTTP call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ScraperClient: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ScraperClient: non-200 response %d: %s", resp.StatusCode, string(respBody))
	}

	var result models.ScraperResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ScraperClient: unmarshal response: %w", err)
	}

	// Tolerant error handling:
	// Only fail hard if scraper errored AND we have no useful data at all.
	// If there's a partial error but we have a screenshot or metrics → proceed with warning.
	if result.Error != "" {
		hasScreenshot := result.ScreenshotB64 != ""
		hasMetrics := len(result.DomMetrics) > 0
		hasSource := result.SourceCode != ""

		if !hasScreenshot && !hasMetrics && !hasSource {
			// Truly nothing — fail hard
			return nil, fmt.Errorf("ScraperClient: scraper failed with no data: %s", result.Error)
		}

		// Partial data — log warning and continue
		log.Printf("[SCRAPER] Warning (partial data): %s | screenshot:%v metrics:%v source:%v",
			result.Error, hasScreenshot, hasMetrics, hasSource)
		result.Error = "" // Clear so it doesn't propagate as fatal
	}

	return &result, nil
}
