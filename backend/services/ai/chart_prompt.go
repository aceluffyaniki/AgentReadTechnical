package ai

// BuildChartSystemPrompt returns the system instruction for chart technical analysis.
func BuildChartSystemPrompt() string {
	return `Kamu adalah AI Trading Analyst profesional spesialis Technical Analysis (TA) dan Smart Money Concepts (SMC).
Kamu WAJIB merespons HANYA dengan JSON murni tanpa markdown, tanpa backtick, tanpa teks di luar JSON.
Analisis harus tegas, berbasis visual chart, dan actionable.
Jika data tidak terlihat di chart, isi "tidak terlihat" atau null. JANGAN mengarang.`
}

// BuildChartUserPrompt creates the text prompt for chart analysis.
func BuildChartUserPrompt(notes, timeframe, source string) string {
	context := ""
	if notes != "" {
		context = "\n\nCATATAN USER: " + notes
	}
	if timeframe == "" {
		timeframe = "tidak diketahui"
	}
	if source == "" {
		source = "tidak diketahui"
	}

	return `Analisis chart trading ini. Timeframe: ` + timeframe + ` | Sumber: ` + source + context + `

Periksa dan isi JSON berikut ini SECARA LENGKAP dari atas ke bawah tanpa skip:

{
  "sinyal_trading": {
    "aksi": "<BUY | SELL | WAIT>",
    "entry": "<harga/zona entry>",
    "tp1": "<target profit 1>",
    "tp2": "<target profit 2>",
    "stop_loss": "<level stop loss>",
    "level_pembatalan": "<harga yang membatalkan seluruh analisa ini>",
    "rr_ratio": "<misal 1:3>",
    "confidence": <0-100>
  },
  "trend": {
    "arah": "<Bullish | Bearish | Ranging>",
    "kekuatan": "<Strong | Moderate | Weak>",
    "struktur": "<HH/HL atau LH/LL, BOS/ChoCH terbaru>",
    "deskripsi": "<ringkasan trend max 30 kata>"
  },
  "pola_teknikal": {
    "pola_terdeteksi": "<nama pola atau 'tidak ada'>",
    "jenis": "<Continuation | Reversal | Candlestick | tidak ada>",
    "status": "<Terbentuk | Dalam Proses | Breakout | Breakdown | tidak ada>",
    "target_pola": "<target harga pola jika ada>",
    "deskripsi": "<max 25 kata>"
  },
  "support_resistance": {
    "support_utama": "<level support terkuat>",
    "resistance_utama": "<level resistance terkuat>",
    "level_sedang_ditest": "<level yang sedang ditest atau 'tidak ada'>",
    "level_lain": "<S/R tambahan>"
  },
  "smart_money": {
    "order_blocks": "<OB bullish/bearish yang terlihat>",
    "fair_value_gaps": "<FVG yang terlihat>",
    "bos_choch": "<BOS atau ChoCH terbaru>",
    "bsl": "<Buy-Side Liquidity: equal highs atau stop loss zona short>",
    "ssl": "<Sell-Side Liquidity: equal lows atau stop loss zona long>",
    "target_likuiditas": "<BSL atau SSL mana yang lebih mungkin diincar>",
    "bias": "<Bullish | Bearish | Ranging>"
  },
  "indikator_dan_divergence": {
    "indikator_terlihat": "<list indikator yang ada di chart>",
    "pembacaan": "<kondisi indikator secara singkat>",
    "divergence_terdeteksi": <true | false>,
    "jenis_divergence": "<Bullish Divergence | Bearish Divergence | Hidden Bullish | Hidden Bearish | tidak ada>",
    "detail_divergence": "<penjelasan singkat atau 'tidak ada'>",
    "korelasi_dengan_price": "<konfirmasi atau kontradiksi price action>"
  },
  "ringkasan": "<max 50 kata, insight paling penting untuk trader>"
}

PENTING: Isi SEMUA field dari atas ke bawah. Kembalikan JSON saja, tanpa teks lain.`
}
