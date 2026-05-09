package ai

import (
	"encoding/json"
	"fmt"

	"github.com/agentrading/backend/models"
)

// BuildChatSystemContext builds the system message for the chat endpoint.
// It injects the full analysis result and contract source as grounding context.
func BuildChatSystemContext(analysis *models.TokenAnalysisDB) (string, error) {
	var aiResult models.AIAnalysis
	if err := json.Unmarshal(analysis.AIResponse, &aiResult); err != nil {
		return "", fmt.Errorf("BuildChatSystemContext: failed to parse AI response: %w", err)
	}

	aiJSON, _ := json.MarshalIndent(aiResult, "", "  ")

	return fmt.Sprintf(`Kamu adalah AgenTrading AI Copilot — asisten riset kripto personal yang cerdas, tajam, dan jujur.

KONTEKS TOKEN AKTIF:
- Address: %s
- Chain: %s
- Waktu Analisis: %s

HASIL ANALISIS LENGKAP:
%s

INSTRUKSI KHUSUS:
- Jawab pertanyaan user SPESIFIK berdasarkan token di atas, bukan secara umum.
- Jika ditanya tentang risiko, berikan penjelasan konkret dari data audit.
- Jika ditanya tentang entry/exit, referensikan angka dari trading_plan di atas.
- Jika ditanya tentang pola chart, referensikan data smart_money di atas.
- Gunakan Bahasa Indonesia yang natural dan mudah dipahami trader retail.
- Boleh berpendapat, tapi selalu sertakan disclaimer bahwa ini bukan financial advice.
- Jika tidak ada informasi yang cukup untuk menjawab, katakan dengan jujur.`,
		analysis.TokenAddress,
		analysis.Chain,
		analysis.AnalyzedAt.Format("2006-01-02 15:04:05 WIB"),
		string(aiJSON),
	), nil
}
