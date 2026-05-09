package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agentrading/backend/models"
)

// BuildAnalyzePrompt builds the multimodal prompt payload for /api/analyze.
// It returns the text portion of the prompt. The image is handled separately
// in vertex_client.go via multipart content.
func BuildAnalyzeSystemPrompt() string {
	return `Kamu adalah AI Crypto Security Auditor dan Trading Analyst profesional.
Kamu WAJIB merespons HANYA dengan JSON murni tanpa markdown, tanpa backtick, tanpa penjelasan di luar JSON.
Analisis kamu harus tegas, akurat, dan berbasis data yang diberikan.
Format output HARUS mengikuti skema JSON yang ditentukan secara PERSIS.`
}

// BuildAnalyzeUserPrompt creates the full text prompt injected alongside the chart screenshot.
func BuildAnalyzeUserPrompt(domMetrics map[string]interface{}, sourceCode, sourceType, chain string) string {
	metricsJSON, _ := json.MarshalIndent(domMetrics, "", "  ")

	contractSection := ""
	if sourceCode != "" {
		lang := "solidity"
		if sourceType == "rust" {
			lang = "rust/SBF"
		}
		contractSection = fmt.Sprintf(`
--- CONTRACT SOURCE CODE (%s) ---
%s
--- END CONTRACT ---
`, lang, sourceCode)
	} else {
		contractSection = "\n--- CONTRACT SOURCE CODE: Tidak tersedia / closed source ---\n"
	}

	return fmt.Sprintf(`Analisis token kripto berikut berdasarkan screenshot chart candlestick yang dilampirkan, metrik on-chain, dan source code kontrak.

CHAIN: %s

--- METRIK ON-CHAIN (dari DOM) ---
%s
--- END METRIK ---
%s
--- INSTRUKSI ANALISIS ---
1. KEAMANAN: Audit source code untuk honeypot, rugpull, blacklist function, mint unlimited, ownership tidak di-renounce.
   Untuk Solana: periksa mint authority, freeze authority, bundle launching, dan pola pump-dump.

2. TRADING PLAN: Berdasarkan screenshot chart candlestick yang kamu lihat, identifikasi:
   - Harga entry ideal berdasarkan support terdekat
   - Target harga jual berdasarkan resistance
   - MCAP saat ini dan target MCAP realistis

3. SMART MONEY CONCEPTS (dari screenshot chart):
   - Identifikasi Order Blocks (bullish/bearish) yang visible
   - Fair Value Gaps (FVG) yang belum terisi
   - Level likuiditas (equal highs/lows, stop hunt zones)
   - Break of Structure (BOS) atau Change of Character (ChoCH) terbaru
   - Market bias keseluruhan

--- OUTPUT JSON YANG WAJIB DIKEMBALIKAN ---
{
  "keamanan": {
    "skor_scam": <integer 0-100, semakin tinggi semakin berbahaya>,
    "status": "<Aman | Risiko Tinggi | Honeypot | Rugpull>",
    "alasan_audit": "<max 30 kata, padat dan jelas>"
  },
  "trading_plan": {
    "estimasi_harga_beli": <float>,
    "estimasi_harga_jual": <float>,
    "mcap_saat_ini": <float>,
    "potensi_mcap_target": <float>,
    "analisa_chart_dan_volume": "<max 50 kata berdasarkan visual chart>"
  },
  "smart_money": {
    "order_blocks": "<deskripsi OB yang teridentifikasi dari chart>",
    "fair_value_gaps": "<FVG yang terlihat dan statusnya>",
    "liquidity_levels": "<level likuiditas kritikal yang terlihat>",
    "bias": "<Bullish | Bearish | Ranging>",
    "smc_summary": "<max 50 kata ringkasan SMC>"
  }
}

PENTING: Kembalikan JSON di atas SAJA. Tidak ada teks lain. Tidak ada markdown. Tidak ada penjelasan.`,
		chain,
		string(metricsJSON),
		contractSection,
	)
}

// ParseAnalysisResponse parses the AI's raw text response into AIAnalysis struct.
// If strict parsing fails, it attempts a best-effort repair for truncated JSON.
func ParseAnalysisResponse(rawResponse string) (*models.AIAnalysis, error) {
	raw := strings.TrimSpace(rawResponse)

	// Strip markdown fences if present
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		var inner []string
		for _, l := range lines {
			if strings.HasPrefix(l, "```") {
				continue
			}
			inner = append(inner, l)
		}
		raw = strings.TrimSpace(strings.Join(inner, "\n"))
	}

	var result models.AIAnalysis
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return &result, nil
	}

	// Attempt repair: close unclosed braces so truncated JSON can be parsed
	repaired := repairJSON(raw)
	if err2 := json.Unmarshal([]byte(repaired), &result); err2 == nil {
		return &result, nil
	}

	// Both attempts failed
	origErr := fmt.Errorf("invalid JSON from AI")
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		origErr = err
	}
	return nil, fmt.Errorf("ParseAnalysisResponse: %w\nRaw: %s", origErr, truncate(raw, 400))
}

// repairJSON attempts to close unclosed braces/brackets in a truncated JSON string.
func repairJSON(s string) string {
	open := 0
	inStr := false
	esc := false
	for _, ch := range s {
		if esc {
			esc = false
			continue
		}
		if ch == '\\' && inStr {
			esc = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if ch == '{' {
			open++
		} else if ch == '}' {
			open--
		}
	}
	// Remove trailing comma before closing
	s = strings.TrimRight(s, " \t\n\r,")
	for i := 0; i < open; i++ {
		s += "}"
	}
	return s
}
