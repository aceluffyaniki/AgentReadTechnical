# 🤖 AgenTrading Copilot

AI-powered crypto trading copilot — analisa token dari Dexscreener dan analisa chart via screenshot menggunakan Gemini AI.

**Fitur:**
- 🔍 **Analisa Token** — Scam detection, Trading Plan, Smart Money Concepts
- 📊 **Analisa Chart** — Upload screenshot / paste clipboard, AI analisa TA lengkap (Sinyal, Trend, Pola TA, S/R, BSL/SSL, Divergence, Invalidation Level)
- 💬 **Chat AI** — Tanya lanjutan tentang hasil analisa token
- 🚀 **One-command launcher** — `python start.py` jalankan semua sekaligus

---

## 🗂️ Struktur Proyek

```
AgenTrading/
├── backend/                    # Go API server (Gin + Gemini SDK)
│   ├── handlers/               # HTTP handlers (analyze, chart, chat)
│   ├── services/ai/            # Gemini AI client + prompts
│   ├── models/                 # Data structures
│   └── main.go
├── scraper/                    # Python scraper microservice (Playwright)
├── frontend_streamlit/         # Streamlit web UI
│   ├── app.py                  # Main UI (multi-tab)
│   └── requirements.txt
├── database/                   # SQL migrations
├── start.py                    # One-command launcher
├── .env.example                # Template environment variables
└── docker-compose.yml
```

---

## ⚙️ Prerequisites (Install Dulu)

| Tool | Versi | Link |
|---|---|---|
| **Go** | ≥ 1.22 | https://go.dev/dl/ |
| **Python** | ≥ 3.11 | https://python.org/downloads/ |
| **PostgreSQL** | ≥ 15 | https://postgresql.org/download/ (atau gunakan DB kantor via Tailscale — lihat bawah) |
| **Git** | latest | https://git-scm.com/ |

---

## 🚀 Setup Step-by-Step

### 1. Clone Repo

```bash
git clone https://github.com/aceluffyaniki/AgentReadTechnical.git
cd AgentReadTechnical
```

### 2. Setup Environment Variables

```bash
# Salin template
copy .env.example .env     # Windows CMD
# cp .env.example .env     # Git Bash / Mac / Linux
```

Buka `.env` dan isi:

| Variable | Wajib | Keterangan |
|---|---|---|
| `DATABASE_URL` | ✅ | Koneksi PostgreSQL (lokal atau via Tailscale) |
| `GEMINI_API_KEY` | ✅ | Dari https://aistudio.google.com/app/apikey |
| `GEMINI_MODEL` | ✅ | `gemini-3.1-pro-preview` (default) |

### 3. Setup Database

**Opsi A — PostgreSQL lokal (install sendiri):**
```sql
-- Buka psql / pgAdmin
CREATE DATABASE agentrading;
```
```
DATABASE_URL=postgres://postgres:PASSWORD@localhost:5432/agentrading?sslmode=disable
```

**Opsi B — Database kantor via Tailscale (tanpa install PostgreSQL):**
```
DATABASE_URL=postgres://postgres:DB_PASSWORD@100.126.178.70:5432/agentrading?sslmode=disable
```
> Syarat: kantor dan rumah terhubung Tailscale, dan PostgreSQL kantor sudah dikonfigurasi untuk terima koneksi Tailscale (lihat bagian **Tailscale DB Setup** di bawah).

**Opsi C — Cloud DB gratis (Neon.tech):**
```
DATABASE_URL=postgres://user:pass@ep-xxx.neon.tech/agentrading?sslmode=require
```

### 4. Setup Backend (Go)

```bash
cd backend
go mod download     # Download semua dependencies
go build ./...      # Verifikasi tidak ada error compile
cd ..
```

### 5. Setup Scraper (Python)

```bash
cd scraper
python -m pip install -r requirements.txt
python -m playwright install chromium    # Install browser untuk scraping
cd ..
```

### 6. Setup Frontend (Streamlit)

```bash
cd frontend_streamlit
python -m pip install -r requirements.txt
cd ..
```

---

## ▶️ Menjalankan Aplikasi

### Cara mudah — 1 command:
```bash
python start.py
```
Tekan `Ctrl+C` untuk menghentikan semua service sekaligus.

### Cara manual — 3 terminal terpisah:

**Terminal 1:**
```bash
cd backend && go run main.go
```
**Terminal 2:**
```bash
cd scraper && python main.py
```
**Terminal 3:**
```bash
cd frontend_streamlit && python -m streamlit run app.py
```

Buka browser → http://localhost:8501

---

## 🌐 Tailscale DB Setup (Koneksi DB Kantor dari Rumah)

Jika ingin pakai PostgreSQL di komputer kantor dari komputer rumah via Tailscale:

**Di komputer KANTOR (lakukan sekali saja):**

1. Temukan file `postgresql.conf` (biasanya di `C:\Program Files\PostgreSQL\15\data\`)
2. Ubah baris `listen_addresses`:
   ```
   listen_addresses = '*'
   ```
3. Buka file `pg_hba.conf` di folder yang sama, tambahkan baris ini di bagian bawah:
   ```
   host    agentrading     postgres        100.64.0.0/10           scram-sha-256
   ```
   > `100.64.0.0/10` adalah subnet Tailscale yang mencakup semua perangkat Tailscale kamu.
4. Restart PostgreSQL service:
   - Buka **Services** Windows → cari **PostgreSQL** → Restart

**Di komputer RUMAH:**

Edit `.env`:
```
DATABASE_URL=postgres://postgres:DB_PASSWORD_KANTOR@100.126.178.70:5432/agentrading?sslmode=disable
```

---

## 📊 Cara Pakai Analisa Chart

1. Buka http://localhost:8501 → Tab **📊 Analisa Chart**
2. **Input gambar** (pilih salah satu):
   - **📋 Paste Clipboard** — Screenshot chart dengan `Win+Shift+S`, lalu klik tombol **"Paste dari Clipboard"** *(cara tercepat!)*
   - **🖼️ Upload File** — Upload file PNG/JPG
   - **🔗 URL Gambar** — Paste URL langsung ke file gambar (bukan halaman web)
3. Pilih **platform** dan **timeframe**
4. Opsional: tambah catatan/konteks
5. Klik **Analisa Chart Sekarang**

**AI akan analisa:**
- Sinyal (BUY / SELL / WAIT) + Entry, TP1, TP2, Stop Loss, R/R
- Trend & Market Structure (HH/HL, BOS, ChoCH)
- Pola Teknikal (Bull Flag, Head & Shoulders, Engulfing, dll.)
- Support & Resistance
- Smart Money Concepts: Order Blocks, FVG, BSL, SSL
- Divergence (true/false + jenis)
- Level Pembatalan Analisa

---

## 🔑 Cara Dapat Gemini API Key (Gratis)

1. Buka https://aistudio.google.com/app/apikey
2. Login dengan akun Google
3. Klik **"Create API Key"**
4. Salin key ke `.env` pada field `GEMINI_API_KEY`

---

## 🐛 Troubleshooting

| Error | Solusi |
|---|---|
| `PostgreSQL connection failed` | PostgreSQL belum jalan, atau `DATABASE_URL` di `.env` salah |
| `no such host: generativelanguage...` | Cek koneksi internet, restart backend |
| `Connection refused :8001` | Pastikan scraper Python sudah jalan (`python start.py`) |
| `model not found` | Pastikan `GEMINI_MODEL=gemini-3.1-pro-preview` di `.env` |
| Paste clipboard tidak muncul | Jalankan `python -m pip install streamlit-paste-button Pillow` |
| Playwright error | Jalankan `python -m playwright install chromium` |
| `start.py` langsung berhenti | Matikan proses lama dulu, baru jalankan `python start.py` |

---

## 📦 Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22, Gin, Google Gen AI SDK (`google.golang.org/genai`) |
| AI Model | Gemini 3.1 Pro Preview (Vision + Text) |
| Scraper | Python, Playwright, FastAPI/Uvicorn |
| Frontend | Streamlit + streamlit-paste-button |
| Database | PostgreSQL 15 |

---

## 🔒 Keamanan

- **JANGAN** commit file `.env` ke GitHub (sudah di-block `.gitignore`)
- Ganti semua API key jika repo di-set public
- Password DB kantor jangan dishare sembarangan
