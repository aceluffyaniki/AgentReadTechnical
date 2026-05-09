package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AnalyzeRequest is what the client sends to /api/analyze
type AnalyzeRequest struct {
	TokenAddress string `json:"token_address" binding:"required"`
	Chain        string `json:"chain" binding:"required,oneof=solana base bsc"`
	ChartURL     string `json:"chart_url" binding:"required"`
}

// AnalyzeResponse is returned to the client after full analysis
type AnalyzeResponse struct {
	AnalysisID uuid.UUID       `json:"analysis_id"`
	TokenAddress string        `json:"token_address"`
	Chain        string        `json:"chain"`
	Result       *AIAnalysis   `json:"result"`
	AnalyzedAt   time.Time     `json:"analyzed_at"`
	TokenUsage   *TokenUsage   `json:"token_usage,omitempty"`
}

// AIAnalysis is the enforced JSON schema returned by the AI
type AIAnalysis struct {
	Keamanan    *SecurityResult `json:"keamanan,omitempty"`
	TradingPlan *TradingPlan    `json:"trading_plan,omitempty"`
	SmartMoney  *SmartMoney     `json:"smart_money,omitempty"`
}

type SecurityResult struct {
	SkorScam    int    `json:"skor_scam"`
	Status      string `json:"status"`
	AlasanAudit string `json:"alasan_audit"`
}

type TradingPlan struct {
	EstimasiBeli          float64 `json:"estimasi_harga_beli"`
	EstimasiJual          float64 `json:"estimasi_harga_jual"`
	McapSaatIni           float64 `json:"mcap_saat_ini"`
	PotensiMcapTarget     float64 `json:"potensi_mcap_target"`
	AnalisaChartDanVolume string  `json:"analisa_chart_dan_volume,omitempty"`
}

type SmartMoney struct {
	OrderBlocks     string `json:"order_blocks,omitempty"`
	FairValueGaps   string `json:"fair_value_gaps,omitempty"`
	LiquidityLevels string `json:"liquidity_levels,omitempty"`
	Bias            string `json:"bias,omitempty"`
	SMCSummary      string `json:"smc_summary,omitempty"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	OutputTokens     int `json:"output_tokens"`
	CacheReadTokens  int `json:"cache_read_tokens"`
}

// TokenAnalysisDB is the database record for token_analysis table
type TokenAnalysisDB struct {
	ID             uuid.UUID        `db:"id"`
	TokenAddress   string           `db:"token_address"`
	Chain          string           `db:"chain"`
	ChartURL       string           `db:"chart_url"`
	ScreenshotB64  string           `db:"screenshot_b64"`
	DomMetrics     json.RawMessage  `db:"dom_metrics"`
	AIResponse     json.RawMessage  `db:"ai_response"`
	SkorScam       int              `db:"skor_scam"`
	StatusKeamanan string           `db:"status_keamanan"`
	SMCBias        string           `db:"smc_bias"`
	ModelUsed      string           `db:"model_used"`
	PromptTokens   int              `db:"prompt_tokens"`
	OutputTokens   int              `db:"output_tokens"`
	CacheReadTokens int             `db:"cache_read_tokens"`
	AnalyzedAt     time.Time        `db:"analyzed_at"`
}

// ScraperResult is what Python scraper returns
type ScraperResult struct {
	ScreenshotB64 string                 `json:"screenshot_b64"`
	DomMetrics    map[string]interface{} `json:"dom_metrics"`
	SourceCode    string                 `json:"source_code"`
	SourceType    string                 `json:"source_type"` // solidity | rust | unknown
	Error         string                 `json:"error,omitempty"`
}
