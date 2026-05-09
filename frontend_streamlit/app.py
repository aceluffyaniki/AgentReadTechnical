"""
AgenTrading — Streamlit MVP Frontend
Multi-tab: Analisa Token | Analisa Chart
"""

import os
import io
import json
import uuid
import base64
import httpx
import streamlit as st
from datetime import datetime
from dotenv import load_dotenv
try:
    from streamlit_paste_button import paste_image_button
    PASTE_AVAILABLE = True
except ImportError:
    PASTE_AVAILABLE = False

load_dotenv("../.env")

BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")

# ─────────────────────────────────────────────────
# Page Config
# ─────────────────────────────────────────────────
st.set_page_config(
    page_title="AgenTrading Copilot",
    page_icon="🤖",
    layout="wide",
    initial_sidebar_state="expanded",
)

# ─────────────────────────────────────────────────
# Custom CSS
# ─────────────────────────────────────────────────
st.markdown("""
<style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap');
    * { font-family: 'Inter', sans-serif; }

    .stApp { background-color: #0a0d14; }
    .block-container { padding-top: 3.5rem !important; }

    /* Hide Streamlit default header/toolbar to reclaim space */
    header[data-testid="stHeader"] { background: rgba(10,13,20,0.95) !important; }
    #MainMenu { visibility: hidden; }
    footer { visibility: hidden; }
    [data-testid="stToolbar"] { display: none; }

    /* Tab styling */
    .stTabs [data-baseweb="tab-list"] {
        gap: 8px;
        background: #0f1319;
        border-radius: 12px;
        padding: 4px 6px;
        border: 1px solid #1e2530;
    }
    .stTabs [data-baseweb="tab"] {
        padding: 8px 20px;
        border-radius: 8px;
        font-weight: 600;
        font-size: 14px;
        color: #8b949e;
    }
    .stTabs [aria-selected="true"] {
        background: linear-gradient(135deg, #1c3a5e, #0d4a8a) !important;
        color: #58a6ff !important;
        border: none !important;
    }

    /* Score Card */
    .score-card {
        border-radius: 12px;
        padding: 16px 20px;
        text-align: center;
        margin-bottom: 12px;
    }
    .score-safe { background: linear-gradient(135deg, #1a3a2a, #0d6e3a); border: 1px solid #2ea55a; }
    .score-risk { background: linear-gradient(135deg, #3a2a0d, #8a5e00); border: 1px solid #f0a500; }
    .score-danger { background: linear-gradient(135deg, #3a0d0d, #8a0000); border: 1px solid #e53935; }
    .score-number { font-size: 56px; font-weight: 800; line-height: 1; }
    .score-label { font-size: 13px; color: #aaa; margin-top: 4px; }
    .status-badge { font-size: 14px; font-weight: 600; margin-top: 8px; }

    /* Metric cards */
    .metric-box {
        background: #0f1319;
        border: 1px solid #1e2530;
        border-radius: 10px;
        padding: 14px;
        text-align: center;
    }
    .metric-label { font-size: 11px; color: #8b949e; text-transform: uppercase; letter-spacing: 0.5px; }
    .metric-value { font-size: 22px; font-weight: 700; color: #e6edf3; margin-top: 4px; }

    /* Chart analysis result cards */
    .ta-card {
        background: #0f1319;
        border: 1px solid #1e2530;
        border-radius: 12px;
        padding: 16px 20px;
        margin-bottom: 12px;
    }
    .ta-card h4 { color: #58a6ff; margin: 0 0 8px 0; font-size: 13px; text-transform: uppercase; letter-spacing: 0.8px; }
    .ta-card p { color: #c9d1d9; margin: 0; font-size: 14px; line-height: 1.6; }

    .signal-buy  { color: #2ea55a; font-weight: 800; font-size: 28px; }
    .signal-sell { color: #ff6b6b; font-weight: 800; font-size: 28px; }
    .signal-wait { color: #f0a500; font-weight: 800; font-size: 28px; }

    .trend-bullish { color: #2ea55a; font-weight: 700; }
    .trend-bearish { color: #ff6b6b; font-weight: 700; }
    .trend-ranging { color: #f0a500; font-weight: 700; }

    .confidence-bar-bg {
        background: #1e2530;
        border-radius: 6px;
        height: 10px;
        width: 100%;
        margin-top: 6px;
    }

    /* Chat bubbles */
    .chat-user {
        background: #1c3a5e;
        border-radius: 12px 12px 4px 12px;
        padding: 10px 14px;
        margin: 6px 0 6px 20%;
        color: #e6edf3;
    }
    .chat-ai {
        background: #0f1319;
        border: 1px solid #1e2530;
        border-radius: 12px 12px 12px 4px;
        padding: 10px 14px;
        margin: 6px 20% 6px 0;
        color: #e6edf3;
    }
    .chat-label { font-size: 10px; color: #8b949e; margin-bottom: 4px; }

    /* SMC badges */
    .smc-badge {
        display: inline-block;
        padding: 3px 10px;
        border-radius: 20px;
        font-size: 12px;
        font-weight: 600;
    }
    .bias-bullish { background: #1a3a2a; color: #2ea55a; border: 1px solid #2ea55a; }
    .bias-bearish { background: #3a0d0d; color: #ff6b6b; border: 1px solid #ff6b6b; }
    .bias-ranging { background: #2a2a10; color: #f0c040; border: 1px solid #f0c040; }
</style>
""", unsafe_allow_html=True)


# ─────────────────────────────────────────────────
# Session State
# ─────────────────────────────────────────────────
if "session_id" not in st.session_state:
    st.session_state.session_id = str(uuid.uuid4())
if "analysis" not in st.session_state:
    st.session_state.analysis = None
if "analysis_id" not in st.session_state:
    st.session_state.analysis_id = None
if "chat_messages" not in st.session_state:
    st.session_state.chat_messages = []
if "screenshot_b64" not in st.session_state:
    st.session_state.screenshot_b64 = None
if "chart_result" not in st.session_state:
    st.session_state.chart_result = None
if "chart_image_preview" not in st.session_state:
    st.session_state.chart_image_preview = None


# ─────────────────────────────────────────────────
# Sidebar
# ─────────────────────────────────────────────────
with st.sidebar:
    st.markdown("### 🤖 AgenTrading Copilot")
    st.markdown("---")
    st.markdown(f"**Session:** `{st.session_state.session_id[:8]}...`")
    if st.button("🔄 Reset Session", use_container_width=True):
        st.session_state.chat_messages = []
        st.session_state.session_id = str(uuid.uuid4())
        st.session_state.analysis = None
        st.session_state.chart_result = None
        st.rerun()
    st.markdown("---")
    st.caption("⚠️ Bukan financial advice. DYOR.")


# ─────────────────────────────────────────────────
# Navigation Tabs
# ─────────────────────────────────────────────────
tab1, tab2 = st.tabs(["🔍 Analisa Token", "📊 Analisa Chart"])


# ═══════════════════════════════════════════════════════
# TAB 1 — Analisa Token (existing feature)
# ═══════════════════════════════════════════════════════
with tab1:
    # Input form
    with st.expander("⚙️ Input Token", expanded=st.session_state.analysis is None):
        col_a, col_b = st.columns([1, 2])
        with col_a:
            chain = st.selectbox("🔗 Chain", ["solana", "base", "bsc"], key="t1_chain")
        with col_b:
            token_address = st.text_input("📍 Token Address", placeholder="Masukkan contract/mint address...", key="t1_addr")

        chart_url = st.text_input("📊 Chart URL (Dexscreener)", placeholder="https://dexscreener.com/...", key="t1_url")
        analyze_btn = st.button("🔍 Analisa Sekarang", use_container_width=True, type="primary", key="t1_analyze")

    if analyze_btn:
        if not token_address or not chart_url:
            st.error("Isi Token Address dan Chart URL dulu!")
        else:
            with st.spinner("🔄 Membuka chart, mengekstrak data, dan memanggil AI..."):
                try:
                    resp = httpx.post(
                        f"{BACKEND_URL}/api/analyze",
                        json={"token_address": token_address, "chain": chain, "chart_url": chart_url},
                        timeout=120.0,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        st.session_state.analysis = data.get("result")
                        st.session_state.analysis_id = data.get("analysis_id")
                        st.session_state.chat_messages = []
                        st.session_state.screenshot_b64 = data.get("screenshot_b64")
                        st.success("✅ Analisis selesai!")
                        st.rerun()
                    else:
                        st.error(f"Error dari backend: {resp.text}")
                except httpx.TimeoutException:
                    st.error("⏱️ Request timeout — scraping atau AI call terlalu lama.")
                except Exception as e:
                    st.error(f"Error: {e}")

    if st.session_state.analysis:
        analysis = st.session_state.analysis
        keamanan = analysis.get("keamanan") or {}
        trading = analysis.get("trading_plan") or {}
        smc = analysis.get("smart_money") or {}

        st.markdown("## 📊 Hasil Analisis Token")
        col_sec, col_chart = st.columns([1, 2])

        with col_sec:
            skor = keamanan.get("skor_scam", 0)
            status = keamanan.get("status", "Unknown")
            card_class = "score-safe"
            score_color = "#2ea55a"
            if skor >= 70:
                card_class = "score-danger"
                score_color = "#ff6b6b"
            elif skor >= 40:
                card_class = "score-risk"
                score_color = "#f0a500"

            st.markdown(f"""
            <div class="score-card {card_class}">
                <div class="score-number" style="color:{score_color}">{skor}</div>
                <div class="score-label">SKOR RISIKO (0=Aman, 100=Berbahaya)</div>
                <div class="status-badge">{status}</div>
            </div>
            """, unsafe_allow_html=True)
            st.markdown("**📋 Audit:**")
            st.info(keamanan.get("alasan_audit", "-"))

        with col_chart:
            if st.session_state.screenshot_b64:
                img_bytes = base64.b64decode(st.session_state.screenshot_b64)
                st.image(img_bytes, caption="📸 Chart Screenshot (Dexscreener)", use_column_width=True)
            else:
                st.markdown("*Screenshot tidak tersedia*")

        st.markdown("---")
        st.markdown("### 💰 Trading Plan")
        m1, m2, m3, m4 = st.columns(4)

        def metric_box(label, value):
            return f"""<div class="metric-box">
                <div class="metric-label">{label}</div>
                <div class="metric-value">{value}</div>
            </div>"""

        m1.markdown(metric_box("Entry (Beli)", f"${trading.get('estimasi_harga_beli', '-')}"), unsafe_allow_html=True)
        m2.markdown(metric_box("Target (Jual)", f"${trading.get('estimasi_harga_jual', '-')}"), unsafe_allow_html=True)
        m3.markdown(metric_box("MCAP Sekarang", f"${trading.get('mcap_saat_ini', '-')}"), unsafe_allow_html=True)
        m4.markdown(metric_box("Target MCAP", f"${trading.get('potensi_mcap_target', '-')}"), unsafe_allow_html=True)

        if trading.get("analisa_chart_dan_volume"):
            st.markdown(f"> **📈 Analisa Chart:** {trading.get('analisa_chart_dan_volume', '-')}")

        if smc:
            st.markdown("---")
            st.markdown("### 🧠 Smart Money Concepts")
            bias = smc.get("bias", "Ranging")
            bias_class = {"Bullish": "bias-bullish", "Bearish": "bias-bearish"}.get(bias, "bias-ranging")
            st.markdown(f'Market Bias: <span class="smc-badge {bias_class}">{bias}</span>', unsafe_allow_html=True)
            if smc.get("smc_summary"):
                st.markdown(f"> {smc.get('smc_summary', '-')}")
            s1, s2, s3 = st.columns(3)
            with s1:
                st.markdown("**📦 Order Blocks**")
                st.caption(smc.get("order_blocks", "-"))
            with s2:
                st.markdown("**🕳️ Fair Value Gaps**")
                st.caption(smc.get("fair_value_gaps", "-"))
            with s3:
                st.markdown("**💧 Liquidity Levels**")
                st.caption(smc.get("liquidity_levels", "-"))

        # Chat
        st.markdown("---")
        st.markdown("### 💬 Tanya AI tentang Token Ini")

        chat_container = st.container()
        with chat_container:
            for msg in st.session_state.chat_messages:
                if msg["role"] == "user":
                    st.markdown(f"""<div class="chat-user"><div class="chat-label">👤 Kamu</div>{msg['content']}</div>""", unsafe_allow_html=True)
                else:
                    st.markdown(f"""<div class="chat-ai"><div class="chat-label">🤖 AI</div>{msg['content']}</div>""", unsafe_allow_html=True)

        with st.form("chat_form", clear_on_submit=True):
            user_input = st.text_area(
                "Pertanyaan kamu:",
                placeholder="Contoh: Kenapa skor scam tinggi? Apa yang mencurigakan?",
                height=80,
            )
            send_btn = st.form_submit_button("📨 Kirim", use_container_width=True)

        if send_btn and user_input.strip():
            st.session_state.chat_messages.append({"role": "user", "content": user_input})
            with st.spinner("AI sedang berpikir..."):
                try:
                    resp = httpx.post(
                        f"{BACKEND_URL}/api/chat",
                        json={
                            "session_id": st.session_state.session_id,
                            "analysis_id": st.session_state.analysis_id,
                            "message": user_input,
                        },
                        timeout=60.0,
                    )
                    if resp.status_code == 200:
                        reply = resp.json().get("reply", "")
                        st.session_state.chat_messages.append({"role": "assistant", "content": reply})
                    else:
                        st.error(f"Chat error: {resp.text}")
                except Exception as e:
                    st.error(f"Error: {e}")
            st.rerun()

    elif not analyze_btn:
        st.markdown("""
        <div style="text-align:center; padding: 60px 20px; color: #8b949e;">
            <h2 style="color:#c9d1d9;">🔍 Analisa Token Crypto</h2>
            <p style="font-size:16px;">Masukkan token address dan chart URL di atas, lalu klik <b>Analisa Sekarang</b></p>
            <p style="font-size:13px; margin-top:16px; color:#6e7681;">
                Deteksi Scam · Trading Plan · Smart Money Concepts · Chat AI Interaktif
            </p>
        </div>
        """, unsafe_allow_html=True)


# ═══════════════════════════════════════════════════════
# TAB 2 — Analisa Chart (new feature)
# ═══════════════════════════════════════════════════════
with tab2:
    st.markdown("## 📊 Analisa Chart")
    st.markdown("Upload screenshot chart atau paste link dari TradingView, Binance, Dexscreener, dll.")

    col_left, col_right = st.columns([1, 1])

    with col_left:
        st.markdown("#### 📤 Input Chart")

        # Source info
        c1, c2 = st.columns(2)
        with c1:
            source = st.selectbox("Platform", ["TradingView", "Binance", "Dexscreener", "Bybit", "OKX", "Lainnya"], key="ch_source")
        with c2:
            timeframe = st.selectbox("Timeframe", ["1m", "5m", "15m", "30m", "1h", "4h", "1d", "1w"], index=4, key="ch_tf")

        # Paste from clipboard
        if PASTE_AVAILABLE:
            paste_result = paste_image_button(
                "📋 Paste dari Clipboard (Ctrl+V)",
                key="ch_paste",
                background_color="#1c3a5e",
                hover_background_color="#0d4a8a",
            )
            if paste_result.image_data is not None:
                buf = io.BytesIO()
                paste_result.image_data.save(buf, format="PNG")
                st.session_state["pasted_image_b64"] = base64.b64encode(buf.getvalue()).decode("utf-8")
                st.session_state["pasted_image_preview"] = buf.getvalue()
                st.success("✅ Gambar dari clipboard berhasil dipaste!")
        else:
            st.caption("`streamlit-paste-button` belum terinstall — jalankan `pip install streamlit-paste-button`")

        st.markdown("<div style='text-align:center; color:#8b949e; margin:8px 0; font-size:13px'>— atau upload file —</div>", unsafe_allow_html=True)

        # Image upload
        uploaded_file = st.file_uploader(
            "🖼️ Upload Screenshot Chart",
            type=["png", "jpg", "jpeg", "webp"],
            help="Screenshot dari TradingView, Binance, Dexscreener, dll.",
            key="ch_upload"
        )

        st.markdown("<div style='text-align:center; color:#8b949e; margin:8px 0; font-size:13px'>— atau URL gambar —</div>", unsafe_allow_html=True)

        # URL input (for direct image URLs — e.g. png/jpg)
        chart_link = st.text_input(
            "🔗 Link Chart (URL gambar langsung)",
            placeholder="https://... (URL langsung ke gambar PNG/JPG)",
            key="ch_url",
            help="Paste URL langsung ke gambar (bukan halaman web). TradingView tidak support direct image URL."
        )

        notes = st.text_area(
            "💬 Catatan / Konteks (opsional)",
            placeholder="Contoh: Ini BTC/USDT, saya bullish tapi takut reversal di resistance 70k...",
            height=80,
            key="ch_notes"
        )

        analyze_chart_btn = st.button("📊 Analisa Chart Sekarang", use_container_width=True, type="primary", key="ch_btn")

    with col_right:
        st.markdown("#### 👁️ Preview")
        pasted_b64 = st.session_state.get("pasted_image_b64", "")
        pasted_preview = st.session_state.get("pasted_image_preview", b"")

        if pasted_preview:
            st.image(pasted_preview, caption="📋 Gambar dari Clipboard", use_column_width=True)
            if st.button("✕ Hapus Paste", key="clear_paste"):
                st.session_state["pasted_image_b64"] = ""
                st.session_state["pasted_image_preview"] = b""
                st.rerun()
        elif uploaded_file:
            st.image(uploaded_file, caption="Screenshot yang akan dianalisa", use_column_width=True)
            uploaded_file.seek(0)
        elif chart_link:
            try:
                st.image(chart_link, caption="Chart dari URL", use_column_width=True)
            except Exception:
                st.warning("Tidak bisa preview URL ini. Pastikan ini URL langsung ke gambar.")
        else:
            st.markdown("""
            <div style="height:200px; display:flex; align-items:center; justify-content:center;
                        background:#0f1319; border:2px dashed #1e2530; border-radius:12px; color:#6e7681; font-size:14px;">
                📊 Preview chart akan muncul di sini
            </div>
            """, unsafe_allow_html=True)

    # Analyze button logic
    if analyze_chart_btn:
        image_b64 = ""
        url_to_send = ""

        pasted_b64 = st.session_state.get("pasted_image_b64", "")
        if pasted_b64:
            # Prioritaskan paste clipboard
            image_b64 = pasted_b64
        elif uploaded_file:
            uploaded_file.seek(0)
            img_bytes = uploaded_file.read()
            image_b64 = base64.b64encode(img_bytes).decode("utf-8")
        elif chart_link:
            url_to_send = chart_link
        else:
            st.error("Paste gambar (📋), upload screenshot, atau masukkan URL chart dulu!")
            st.stop()

        with st.spinner("🤖 AI sedang menganalisa chart..."):
            try:
                resp = httpx.post(
                    f"{BACKEND_URL}/api/analyze-chart",
                    json={
                        "image_b64": image_b64,
                        "chart_url": url_to_send,
                        "notes": notes,
                        "timeframe": timeframe,
                        "source": source,
                    },
                    timeout=120.0,
                )
                if resp.status_code == 200:
                    data = resp.json()
                    st.session_state.chart_result = data.get("result", {})
                    st.success("✅ Analisis chart selesai!")
                    st.rerun()
                else:
                    st.error(f"Error: {resp.text}")
            except httpx.TimeoutException:
                st.error("⏱️ Timeout — coba gambar lebih kecil.")
            except Exception as e:
                st.error(f"Error: {e}")

    # Display chart analysis result
    if st.session_state.chart_result:
        r = st.session_state.chart_result
        trend    = r.get("trend") or {}
        pola     = r.get("pola_teknikal") or {}
        sr       = r.get("support_resistance") or {}
        smc      = r.get("smart_money") or {}
        sinyal   = r.get("sinyal_trading") or {}
        indikdiv = r.get("indikator_dan_divergence") or {}
        ringkasan = r.get("ringkasan", "")

        st.markdown("---")
        st.markdown("### 🎯 Hasil Analisa Chart AI")

        # ── Row 1: Sinyal · Trend · Confidence ─────────────────────
        c1, c2, c3 = st.columns(3)

        with c1:
            aksi = sinyal.get("aksi", "WAIT").upper()
            aksi_class = {"BUY": "signal-buy", "SELL": "signal-sell"}.get(aksi, "signal-wait")
            emoji = {"BUY": "🟢", "SELL": "🔴", "WAIT": "🟡"}.get(aksi, "🟡")
            st.markdown(f"""
            <div class="ta-card" style="text-align:center;">
                <h4>SINYAL</h4>
                <div class="{aksi_class}">{emoji} {aksi}</div>
                <div style="color:#8b949e; font-size:12px; margin-top:6px;">R/R: {sinyal.get('rr_ratio', '-')}</div>
            </div>""", unsafe_allow_html=True)

        with c2:
            arah = trend.get("arah", "Ranging")
            kekuatan = trend.get("kekuatan", "-")
            trend_class = {"Bullish": "trend-bullish", "Bearish": "trend-bearish"}.get(arah, "trend-ranging")
            st.markdown(f"""
            <div class="ta-card" style="text-align:center;">
                <h4>TREND</h4>
                <div class="{trend_class}" style="font-size:22px; font-weight:700;">{arah}</div>
                <div style="color:#8b949e; font-size:12px; margin-top:6px;">{kekuatan}</div>
            </div>""", unsafe_allow_html=True)

        with c3:
            conf = sinyal.get("confidence", 0)
            if isinstance(conf, str):
                conf = 0
            conf_color = "#2ea55a" if conf >= 70 else ("#f0a500" if conf >= 40 else "#ff6b6b")
            st.markdown(f"""
            <div class="ta-card" style="text-align:center;">
                <h4>CONFIDENCE</h4>
                <div style="font-size:36px; font-weight:800; color:{conf_color};">{conf}%</div>
                <div class="confidence-bar-bg">
                    <div style="background:{conf_color}; height:10px; border-radius:6px; width:{conf}%;"></div>
                </div>
            </div>""", unsafe_allow_html=True)

        # ── Row 2: Pola Teknikal ────────────────────────────────────
        if pola.get("pola_terdeteksi") and pola.get("pola_terdeteksi") != "tidak ada pola terkonfirmasi":
            st.markdown("---")
            st.markdown("#### 📐 Pola Teknikal")
            status_p = pola.get("status", "-")
            jenis_p  = pola.get("jenis", "-")
            target_p = pola.get("target_pola", "-")
            status_color = {
                "Breakout": "#2ea55a", "Breakdown": "#ff6b6b",
                "Terbentuk": "#58a6ff", "Dalam Proses": "#f0a500"
            }.get(status_p, "#c9d1d9")
            pola_emoji = {
                "Continuation": "🔄", "Reversal": "🔀", "Candlestick": "🕯️"
            }.get(jenis_p, "📊")
            st.markdown(f"""
            <div class="ta-card">
                <h4>POLA TERDETEKSI</h4>
                <div style="display:flex; align-items:center; gap:12px; flex-wrap:wrap;">
                    <span style="font-size:22px; font-weight:700; color:#e6edf3;">{pola_emoji} {pola.get('pola_terdeteksi', '-')}</span>
                    <span class="smc-badge" style="background:#1e2530; color:{status_color}; border:1px solid {status_color};">{status_p}</span>
                    <span style="color:#8b949e; font-size:12px;">{jenis_p}</span>
                </div>
                <p style="margin-top:8px; font-size:13px; color:#c9d1d9;">{pola.get('deskripsi', '-')}</p>
                {f'<p style="color:#f0a500; font-size:13px; margin-top:4px;">🎯 Target Pola: <b>{target_p}</b></p>' if target_p and target_p != "tidak ada" else ""}
            </div>""", unsafe_allow_html=True)

        # ── Row 3: Trading Levels + Invalidation ───────────────────
        st.markdown("---")
        st.markdown("#### 💰 Trading Levels")

        e1, e2, e3, e4 = st.columns(4)
        def tcard(label, value, color="#e6edf3"):
            return f"""<div class="metric-box">
                <div class="metric-label">{label}</div>
                <div class="metric-value" style="color:{color}; font-size:15px;">{value}</div>
            </div>"""

        e1.markdown(tcard("Entry", sinyal.get("entry", "-"), "#58a6ff"), unsafe_allow_html=True)
        e2.markdown(tcard("TP 1", sinyal.get("tp1", "-"), "#2ea55a"), unsafe_allow_html=True)
        e3.markdown(tcard("TP 2", sinyal.get("tp2", "-"), "#2ea55a"), unsafe_allow_html=True)
        e4.markdown(tcard("Stop Loss", sinyal.get("stop_loss", "-"), "#ff6b6b"), unsafe_allow_html=True)

        # Invalidation level — kotak merah penting
        level_batal = sinyal.get("level_pembatalan", "")
        if level_batal and level_batal not in ("-", "tidak ada", ""):
            st.markdown(f"""
            <div style="background:#2d0f0f; border:1px solid #8a0000; border-left:4px solid #e53935;
                        border-radius:8px; padding:12px 16px; margin-top:10px;">
                <span style="color:#ff6b6b; font-size:12px; font-weight:700; text-transform:uppercase; letter-spacing:0.8px;">
                    ⚠️ Level Pembatalan Analisa
                </span>
                <div style="color:#ffcdd2; font-size:15px; margin-top:4px; font-weight:600;">{level_batal}</div>
                <div style="color:#ef9a9a; font-size:12px; margin-top:2px;">Jika harga mencapai level ini, bias di atas sepenuhnya dianggap salah.</div>
            </div>""", unsafe_allow_html=True)

        # ── Row 4: Support/Resistance + SMC + Liquidity ────────────
        st.markdown("---")
        col_sr, col_smc = st.columns(2)

        with col_sr:
            st.markdown("#### 📐 Support & Resistance")
            level_test = sr.get("level_sedang_ditest", "")
            st.markdown(f"""
            <div class="ta-card">
                <h4>SUPPORT UTAMA</h4>
                <p style="color:#2ea55a; font-size:16px; font-weight:600;">{sr.get('support_utama', '-')}</p>
            </div>
            <div class="ta-card">
                <h4>RESISTANCE UTAMA</h4>
                <p style="color:#ff6b6b; font-size:16px; font-weight:600;">{sr.get('resistance_utama', '-')}</p>
            </div>
            {f'<div class="ta-card"><h4>⚡ SEDANG DITEST</h4><p style="color:#f0a500; font-weight:600;">{level_test}</p></div>' if level_test and level_test != "tidak ada" else ""}
            <div class="ta-card">
                <h4>LEVEL LAIN</h4><p>{sr.get('level_lain', '-')}</p>
            </div>""", unsafe_allow_html=True)

        with col_smc:
            st.markdown("#### 🧠 Smart Money Concepts")
            bias = smc.get("bias", "Ranging")
            bias_class = {"Bullish": "bias-bullish", "Bearish": "bias-bearish"}.get(bias, "bias-ranging")
            st.markdown(f'Market Bias: <span class="smc-badge {bias_class}">{bias}</span>', unsafe_allow_html=True)
            bsl = smc.get("bsl", "-")
            ssl = smc.get("ssl", "-")
            target_liq = smc.get("target_likuiditas", "-")
            st.markdown(f"""
            <div class="ta-card" style="margin-top:12px;">
                <h4>ORDER BLOCKS</h4><p>{smc.get('order_blocks', '-')}</p>
            </div>
            <div class="ta-card">
                <h4>FAIR VALUE GAPS</h4><p>{smc.get('fair_value_gaps', '-')}</p>
            </div>
            <div class="ta-card" style="border-color:#1a3a2a;">
                <h4 style="color:#2ea55a;">🟢 BSL (Buy-Side Liquidity)</h4><p>{bsl}</p>
            </div>
            <div class="ta-card" style="border-color:#3a0d0d;">
                <h4 style="color:#ff6b6b;">🔴 SSL (Sell-Side Liquidity)</h4><p>{ssl}</p>
            </div>
            <div class="ta-card" style="border-color:#1c3a5e;">
                <h4 style="color:#58a6ff;">🎯 Target Likuiditas Selanjutnya</h4><p style="font-weight:600;">{target_liq}</p>
            </div>""", unsafe_allow_html=True)

        # ── Row 5: Indikator & Divergence ──────────────────────────
        st.markdown("---")
        st.markdown("#### 📈 Indikator & Divergence")

        div_terdeteksi = indikdiv.get("divergence_terdeteksi", False)
        jenis_div = indikdiv.get("jenis_divergence", "tidak ada")
        detail_div = indikdiv.get("detail_divergence", "-")
        korelasi = indikdiv.get("korelasi_dengan_price", "-")
        indikator_list = indikdiv.get("indikator_terlihat", "-")
        pembacaan = indikdiv.get("pembacaan", "-")

        col_ind, col_div = st.columns(2)
        with col_ind:
            st.markdown(f"""
            <div class="ta-card">
                <h4>INDIKATOR TERDETEKSI</h4>
                <p style="color:#58a6ff; font-weight:600;">{indikator_list}</p>
                <p style="margin-top:6px; font-size:13px;">{pembacaan}</p>
            </div>
            <div class="ta-card">
                <h4>KORELASI DENGAN PRICE ACTION</h4>
                <p style="font-size:13px;">{korelasi}</p>
            </div>""", unsafe_allow_html=True)

        with col_div:
            if div_terdeteksi:
                jenis_div_str = str(jenis_div) if jenis_div else ""
                is_bearish = "Bearish" in jenis_div_str
                div_color = "#ff6b6b" if is_bearish else "#2ea55a"
                div_icon = "⚠️" if is_bearish else "✅"
                bg_color = "#2d0f0f" if is_bearish else "#0d2a1a"
                st.markdown(f"""
                <div class="ta-card" style="border-color:{div_color}; background:{bg_color};">
                    <h4 style="color:{div_color};">DIVERGENCE TERDETEKSI {div_icon}</h4>
                    <div style="color:{div_color}; font-size:16px; font-weight:700; margin-bottom:8px;">{jenis_div}</div>
                    <p style="font-size:13px;">{detail_div}</p>
                </div>""", unsafe_allow_html=True)
            else:
                st.markdown(f"""
                <div class="ta-card">
                    <h4>DIVERGENCE</h4>
                    <div style="color:#8b949e; font-size:15px;">✓ Tidak Terdeteksi</div>
                    <p style="font-size:12px; color:#6e7681; margin-top:4px;">Indikator dan price action searah.</p>
                </div>""", unsafe_allow_html=True)

            # Struktur pasar
            struktur = trend.get("struktur", "")
            if struktur:
                st.markdown(f"""
                <div class="ta-card" style="margin-top:8px;">
                    <h4>MARKET STRUCTURE</h4>
                    <p style="font-size:13px;">{struktur}</p>
                </div>""", unsafe_allow_html=True)

        # ── Ringkasan ───────────────────────────────────────────────
        st.markdown("---")
        if ringkasan:
            st.markdown("#### 📝 Ringkasan AI")
            st.info(ringkasan)

        if st.button("🗑️ Clear Hasil", key="clear_chart"):
            st.session_state.chart_result = None
            st.rerun()


    elif not analyze_chart_btn:
        st.markdown("""
        <div style="text-align:center; padding: 40px 20px; color: #8b949e; margin-top:24px;">
            <p style="font-size:48px; margin:0;">📊</p>
            <h3 style="color:#c9d1d9; margin:12px 0 8px;">Analisa Chart dengan AI</h3>
            <p style="font-size:14px;">Upload screenshot dari TradingView, Binance, Dexscreener, dll.<br>
            AI akan membaca chart dan memberikan analisa TA lengkap.</p>
        </div>
        """, unsafe_allow_html=True)
