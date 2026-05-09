package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	BackendPort string
	GinMode     string

	// Database
	DatabaseURL string

	// Scraper microservice
	ScraperURL string

	// Vertex AI
	GoogleCredentials string
	VertexProjectID   string
	VertexLocation    string
	VertexModel       string

	// Explorer APIs
	BscScanAPIKey  string
	BaseScanAPIKey string
	HeliusAPIKey   string
	SolscanAPIKey  string

	// Telegram
	TelegramBotToken string
	TelegramChatID   string

	// Cost Guard
	MaxPromptTokens    int
	ScamAlertThreshold int
}

var AppConfig Config

func Load() {
	// Load .env file (ignore error if not found, rely on system env)
	_ = godotenv.Load("../.env")

	AppConfig = Config{
		BackendPort:        getEnv("BACKEND_PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/agentrading"),
		ScraperURL:         getEnv("SCRAPER_URL", "http://localhost:8001"),
		GoogleCredentials:  getEnv("GOOGLE_APPLICATION_CREDENTIALS", "./service-account.json"),
		VertexProjectID:    getEnv("VERTEX_PROJECT_ID", ""),
		VertexLocation:     getEnv("VERTEX_LOCATION", "us-east5"),
		VertexModel:        getEnv("VERTEX_MODEL", "claude-opus-4-5"),
		BscScanAPIKey:      getEnv("BSCSCAN_API_KEY", ""),
		BaseScanAPIKey:     getEnv("BASESCAN_API_KEY", ""),
		HeliusAPIKey:       getEnv("HELIUS_API_KEY", ""),
		SolscanAPIKey:      getEnv("SOLSCAN_API_KEY", ""),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     getEnv("TELEGRAM_CHAT_ID", ""),
		MaxPromptTokens:    getEnvInt("MAX_PROMPT_TOKENS", 20000),
		ScamAlertThreshold: getEnvInt("SCAM_ALERT_THRESHOLD", 70),
	}

	if AppConfig.VertexProjectID == "" {
		log.Println("[WARN] VERTEX_PROJECT_ID not set — AI calls will fail")
	}

	log.Printf("[CONFIG] Loaded — Port:%s Model:%s Chain support: Solana/Base/BSC",
		AppConfig.BackendPort, AppConfig.VertexModel)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
