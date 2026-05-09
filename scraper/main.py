"""
main.py — FastAPI microservice exposing /scrape endpoint.
Called by Go backend to get DOM metrics, chart screenshot, and contract source.
"""

import os
import sys
import asyncio
import traceback
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from scraper import scrape_dexscreener
from explorer_client import get_contract_source
from chain_config import CHAIN_CONFIG

load_dotenv("../.env")

# Force UTF-8 output on Windows
if sys.platform == "win32":
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")

app = FastAPI(
    title="AgenTrading Scraper Microservice",
    description="Playwright-based scraper for Dexscreener chart data and contract source",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global exception handler — returns JSON instead of plain text 500
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    tb = traceback.format_exc()
    print(f"[ERROR] Unhandled exception:\n{tb}")
    return JSONResponse(
        status_code=500,
        content={"error": str(exc), "traceback": tb},
    )


class ScrapeRequest(BaseModel):
    chain: str          # solana | base | bsc
    token_address: str
    chart_url: str


class ScrapeResponse(BaseModel):
    screenshot_b64: str = ""
    dom_metrics: dict = {}
    source_code: str = ""
    source_type: str = "unknown"
    error: str = ""


@app.get("/health")
async def health():
    return {"status": "ok", "service": "agentrading-scraper"}


@app.post("/scrape", response_model=ScrapeResponse)
async def scrape(req: ScrapeRequest):
    if req.chain not in CHAIN_CONFIG:
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported chain: {req.chain}. Supported: {list(CHAIN_CONFIG.keys())}"
        )

    print(f"[API] /scrape — chain={req.chain} token={req.token_address[:12]}...")

    chain_cfg = CHAIN_CONFIG[req.chain]
    explorer_type = chain_cfg["explorer_type"]

    # Run independently so one failure doesn't crash the other
    scrape_result = None
    source_code = ""
    source_type = "unknown"
    error_msg = ""

    # Step 1: Scrape chart (Playwright)
    try:
        scrape_result = await asyncio.wait_for(
            scrape_dexscreener(req.chart_url, req.chain, req.token_address),
            timeout=60.0
        )
        error_msg = scrape_result.get("error") or ""
    except asyncio.TimeoutError:
        error_msg = "Playwright scraping timed out after 60s"
        print(f"[ERROR] {error_msg}")
        scrape_result = {"screenshot_b64": "", "dom_metrics": {}}
    except Exception as e:
        error_msg = f"Playwright error: {e}"
        print(f"[ERROR] {error_msg}\n{traceback.format_exc()}")
        scrape_result = {"screenshot_b64": "", "dom_metrics": {}}

    # Step 2: Fetch contract source (explorer API)
    try:
        source_code, source_type = await asyncio.wait_for(
            get_contract_source(req.token_address, req.chain, explorer_type),
            timeout=15.0
        )
    except asyncio.TimeoutError:
        print("[WARN] Explorer API timed out after 15s, proceeding without source code")
    except Exception as e:
        print(f"[WARN] Explorer fetch error: {e}")

    print(f"[API] Done — screenshot:{bool(scrape_result.get('screenshot_b64'))} source_type:{source_type} error:{error_msg or 'none'}")

    return ScrapeResponse(
        screenshot_b64=scrape_result.get("screenshot_b64", ""),
        dom_metrics=scrape_result.get("dom_metrics", {}),
        source_code=source_code,
        source_type=source_type,
        error=error_msg,
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8001, reload=True)
