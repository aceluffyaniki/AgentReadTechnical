package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/agentrading/backend/config"
	"github.com/agentrading/backend/handlers"
	"github.com/agentrading/backend/middleware"
	"github.com/agentrading/backend/repository"
	"github.com/agentrading/backend/services"
	"github.com/agentrading/backend/services/ai"
)

func main() {
	// ── 1. Load configuration ──────────────────────────────────────
	config.Load()
	gin.SetMode(config.AppConfig.GinMode)

	// ── 2. Connect to PostgreSQL (non-fatal — app runs without DB) ────────────
	if err := repository.Connect(); err != nil {
		log.Printf("[WARN] PostgreSQL unavailable — DB features disabled: %v", err)
		log.Printf("[WARN] Analisa Chart tetap berjalan. Analisa Token tidak bisa simpan history.")
	} else {
		defer repository.Close()
	}

	// ── 3. Initialize Vertex AI client ────────────────────────────
	ctx := context.Background()
	aiClient, err := ai.NewVertexClient(ctx)
	if err != nil {
		log.Fatalf("[FATAL] Vertex AI initialization failed: %v", err)
	}

	// ── 4. Initialize services ────────────────────────────────────
	scraperClient := services.NewScraperClient()

	// ── 5. Initialize handlers ────────────────────────────────────
	analyzeHandler := handlers.NewAnalyzeHandler(scraperClient, aiClient)
	chatHandler := handlers.NewChatHandler(aiClient)
	chartHandler := handlers.NewChartAnalyzeHandler(aiClient)

	// ── 6. Setup Gin router ───────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// ── 7. Routes ─────────────────────────────────────────────────
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "agentrading-backend",
				"time":    time.Now().Format(time.RFC3339),
			})
		})

		// Core analysis endpoint
		api.POST("/analyze", analyzeHandler.Handle)

		// Chart TA analysis endpoint
		api.POST("/analyze-chart", chartHandler.Handle)

		// Chat endpoints
		api.POST("/chat", chatHandler.Handle)
		api.GET("/chat/history", chatHandler.GetHistory)
	}

	// ── 8. Start server with graceful shutdown ────────────────────
	srv := &http.Server{
		Addr:         ":" + config.AppConfig.BackendPort,
		Handler:      r,
		ReadTimeout:  120 * time.Second, // Long timeout for AI calls
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		log.Printf("[SERVER] AgenTrading Backend starting on port %s", config.AppConfig.BackendPort)
		log.Printf("[SERVER] Model: %s | Scraper: %s", config.AppConfig.VertexModel, config.AppConfig.ScraperURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER] Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[WARN] Server shutdown error: %v", err)
	}
	log.Println("[SERVER] Goodbye.")
}
