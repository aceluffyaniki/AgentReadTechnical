package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/agentrading/backend/models"
)

// VertexClient keeps its original struct name so we don't break existing handlers,
// but now it wraps the official Google Gen AI Go SDK client.
type VertexClient struct {
	client *genai.Client
	model  string
}

// NewVertexClient creates a Google Gen AI SDK client using GEMINI_API_KEY.
func NewVertexClient(ctx context.Context) (*VertexClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is missing in .env")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-pro-preview-03-25"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gen AI client: %w", err)
	}

	log.Printf("[AI] Mode: Google Gen AI SDK — model:%s", model)

	return &VertexClient{
		client: client,
		model:  model,
	}, nil
}

// ─────────────────────────────────── AnalyzeToken ────────────────────────────

func (v *VertexClient) AnalyzeToken(
	ctx context.Context,
	screenshotB64 string,
	userPromptText string,
) (*models.AIAnalysis, *models.TokenUsage, error) {

	var parts []*genai.Part

	// Add screenshot image if provided
	if screenshotB64 != "" {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: "image/png",
				Data:     []byte(screenshotB64),
			},
		})
	}

	parts = append(parts, genai.NewPartFromText(userPromptText))

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(BuildAnalyzeSystemPrompt(), genai.RoleUser),
		Temperature:       genai.Ptr[float32](0.1),
		MaxOutputTokens:   4000,
		ResponseMIMEType:  "application/json",
	}

	resp, err := v.client.Models.GenerateContent(ctx, v.model, contents, config)
	if err != nil {
		return nil, nil, fmt.Errorf("AnalyzeToken: %w", err)
	}

	rawText := resp.Text()
	log.Printf("[AI] Raw response: %s", truncate(rawText, 150))

	analysis, err := ParseAnalysisResponse(rawText)
	if err != nil {
		return nil, nil, err
	}

	usage := &models.TokenUsage{}
	if resp.UsageMetadata != nil {
		usage.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}
	log.Printf("[AI] Done — input:%d output:%d tokens", usage.PromptTokens, usage.OutputTokens)
	return analysis, usage, nil
}

// ─────────────────────────────────── Chat ────────────────────────────────────

func (v *VertexClient) Chat(
	ctx context.Context,
	systemContext string,
	history []models.ChatMessage,
	userMessage string,
) (string, int, error) {

	var contents []*genai.Content

	// Convert history format: Gemini uses "user" and "model"
	for _, msg := range history {
		var role genai.Role = genai.RoleUser
		if msg.Role == "assistant" {
			role = genai.RoleModel
		}
		contents = append(contents, genai.NewContentFromText(msg.Content, role))
	}

	// Add new message
	contents = append(contents, genai.NewContentFromText(userMessage, genai.RoleUser))

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemContext, genai.RoleUser),
		Temperature:       genai.Ptr[float32](0.7),
		MaxOutputTokens:   2000,
	}

	resp, err := v.client.Models.GenerateContent(ctx, v.model, contents, config)
	if err != nil {
		return "", 0, fmt.Errorf("Chat: %w", err)
	}

	reply := resp.Text()
	outputTokens := 0
	if resp.UsageMetadata != nil {
		outputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}
	return reply, outputTokens, nil
}

// ─────────────────────────────────── HealthCheck ─────────────────────────────

func (v *VertexClient) HealthCheck(ctx context.Context) error {
	contents := []*genai.Content{
		genai.NewContentFromText("ping", genai.RoleUser),
	}
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 5,
	}
	_, err := v.client.Models.GenerateContent(ctx, v.model, contents, config)
	return err
}

// ─────────────────────────────────── AnalyzeChart ────────────────────────────

// AnalyzeChart sends a chart image (base64) or fetches a URL screenshot,
// then returns a rich TA analysis as a raw map (flexible JSON).
func (v *VertexClient) AnalyzeChart(
	ctx context.Context,
	imageB64 string,
	chartURL string,
	userPromptText string,
) (map[string]interface{}, *models.TokenUsage, error) {

	var parts []*genai.Part

	if imageB64 != "" {
		// Decode base64 — strip data-URL prefix if present
		raw := imageB64
		if idx := len("data:image/png;base64,"); len(raw) > idx && raw[:idx] == "data:image/png;base64," {
			raw = raw[idx:]
		} else if idx2 := len("data:image/jpeg;base64,"); len(raw) > idx2 && raw[:idx2] == "data:image/jpeg;base64," {
			raw = raw[idx2:]
		}

		imgBytes, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("AnalyzeChart: invalid base64 image: %w", err)
		}

		// Detect mime type (default PNG)
		mimeType := "image/png"
		if len(imgBytes) > 2 && imgBytes[0] == 0xFF && imgBytes[1] == 0xD8 {
			mimeType = "image/jpeg"
		} else if len(imgBytes) > 8 && string(imgBytes[1:4]) == "PNG" {
			mimeType = "image/png"
		}

		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     imgBytes,
			},
		})
	} else if chartURL != "" {
		// Download image from URL — Gemini FileData only supports GCS URIs,
		// so we must fetch the bytes manually.
		imgBytes, mimeType, err := fetchImageFromURL(ctx, chartURL)
		if err != nil {
			return nil, nil, fmt.Errorf("AnalyzeChart: failed to fetch chart URL: %w", err)
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     imgBytes,
			},
		})
	}

	parts = append(parts, genai.NewPartFromText(userPromptText))

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(BuildChartSystemPrompt(), genai.RoleUser),
		Temperature:       genai.Ptr[float32](0.2),
		MaxOutputTokens:   4000,
		// Note: do NOT use ResponseMIMEType:"application/json" here — it causes
		// gemini-3.1-pro-preview to terminate output early before all JSON fields are generated.
	}

	resp, err := v.client.Models.GenerateContent(ctx, v.model, contents, config)
	if err != nil {
		return nil, nil, fmt.Errorf("AnalyzeChart: %w", err)
	}

	rawText := resp.Text()
	log.Printf("[AI] Chart raw response (%d chars):\n%s", len(rawText), truncate(rawText, 3000))

	// Parse into generic map — strip markdown fences if model wrapped output
	cleaned := extractJSON(rawText)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		// Try repair for truncated JSON
		repaired := repairJSON(cleaned)
		if err2 := json.Unmarshal([]byte(repaired), &result); err2 != nil {
			return nil, nil, fmt.Errorf("AnalyzeChart: invalid JSON: %w\nRaw: %s", err, truncate(rawText, 500))
		}
		log.Printf("[AI] Chart JSON was repaired (truncated response)")
	}

	usage := &models.TokenUsage{}
	if resp.UsageMetadata != nil {
		usage.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}
	log.Printf("[AI] Chart done — input:%d output:%d tokens", usage.PromptTokens, usage.OutputTokens)
	return result, usage, nil
}

// ─────────────────────────────────── Helpers ─────────────────────────────────

// fetchImageFromURL downloads an image from an HTTP(S) URL and returns its bytes + mime type.
func fetchImageFromURL(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d fetching image URL", resp.StatusCode)
	}

	imgBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // max 10MB
	if err != nil {
		return nil, "", err
	}

	// Detect MIME type from Content-Type header or magic bytes
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		if len(imgBytes) > 2 && imgBytes[0] == 0xFF && imgBytes[1] == 0xD8 {
			mimeType = "image/jpeg"
		} else {
			mimeType = "image/png"
		}
	}
	// Strip parameters (e.g. "image/png; charset=utf-8")
	if idx := len(mimeType); idx > 0 {
		for i, c := range mimeType {
			if c == ';' {
				mimeType = mimeType[:i]
				break
			}
		}
	}

	return imgBytes, mimeType, nil
}

// extractJSON strips markdown fences and returns the first JSON object found in s.
// E.g. strips ```json ... ``` wrappers that models sometimes emit without ResponseMIMEType.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		// Find end of first line
		end := strings.Index(s, "\n")
		if end == -1 {
			return s
		}
		s = s[end+1:]
		// Strip trailing ```
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	// Find first { and last }
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
