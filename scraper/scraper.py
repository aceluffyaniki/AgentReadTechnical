"""
scraper.py — Playwright-based DOM extractor and chart screenshot tool.
Uses async_playwright().start() pattern for compatibility inside uvicorn event loop.
"""

import asyncio
import base64
import traceback
from typing import Dict, Any

from playwright.async_api import async_playwright, Page, Browser
from chain_config import CHAIN_CONFIG


async def scrape_dexscreener(
    chart_url: str,
    chain: str,
    token_address: str,
) -> Dict[str, Any]:
    """
    Opens chart_url with Playwright, extracts DOM metrics and screenshots chart.
    Uses async_playwright().start() instead of context manager for uvicorn compatibility.
    """
    result: Dict[str, Any] = {
        "screenshot_b64": "",
        "dom_metrics": {},
        "error": None,
    }

    playwright = await async_playwright().start()
    try:
        browser: Browser = await playwright.chromium.launch(
            headless=True,
            args=[
                "--no-sandbox",
                "--disable-setuid-sandbox",
                "--disable-blink-features=AutomationControlled",
                "--window-size=1440,900",
            ],
        )

        context = await browser.new_context(
            viewport={"width": 1440, "height": 900},
            user_agent=(
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                "AppleWebKit/537.36 (KHTML, like Gecko) "
                "Chrome/120.0.0.0 Safari/537.36"
            ),
        )

        page: Page = await context.new_page()

        try:
            print(f"[SCRAPER] Opening: {chart_url}")
            await page.goto(chart_url, wait_until="domcontentloaded", timeout=45000)
            await page.wait_for_timeout(5000)

            # Extract DOM metrics (non-fatal if fails)
            try:
                metrics = await _extract_metrics(page)
                result["dom_metrics"] = metrics
                print(f"[SCRAPER] Metrics: {list(metrics.keys())}")
            except Exception as me:
                print(f"[SCRAPER] Metrics extraction error: {repr(me)}")

            # Always take a full-page screenshot as base fallback first
            try:
                img_bytes = await page.screenshot(full_page=False)
                result["screenshot_b64"] = base64.b64encode(img_bytes).decode("utf-8")
                print(f"[SCRAPER] Base screenshot: {len(result['screenshot_b64'])} chars")
            except Exception as se:
                print(f"[SCRAPER] Base screenshot failed: {repr(se)}")

            # Try to improve screenshot by targeting chart element
            try:
                screenshot_b64 = await _take_chart_screenshot(page)
                if screenshot_b64:
                    result["screenshot_b64"] = screenshot_b64
                    print(f"[SCRAPER] Chart screenshot: {len(screenshot_b64)} chars")
            except Exception as ce:
                print(f"[SCRAPER] Chart screenshot failed (using base): {repr(ce)}")

        except Exception as e:
            err_msg = repr(e) if not str(e) else str(e)
            print(f"[SCRAPER] Page navigation error: {err_msg}")
            result["error"] = err_msg

            # Fallback: try screenshot even after navigation error
            try:
                img_bytes = await page.screenshot(full_page=False)
                result["screenshot_b64"] = base64.b64encode(img_bytes).decode("utf-8")
                print("[SCRAPER] Post-error fallback screenshot taken")
            except Exception as e2:
                print(f"[SCRAPER] Fallback screenshot also failed: {repr(e2)}")

        finally:
            await context.close()
            await browser.close()

    except Exception as e:
        err_msg = repr(e) if not str(e) else str(e)
        print(f"[SCRAPER] Browser launch error: {err_msg}")
        result["error"] = err_msg
    finally:
        await playwright.stop()

    return result



async def _extract_metrics(page: Page) -> Dict[str, Any]:
    """Extract price, mcap, volume, liquidity from Dexscreener DOM."""
    metrics: Dict[str, Any] = {}

    # Try waiting for stats to appear
    try:
        await page.wait_for_selector(
            "[class*='statsValue'], [class*='priceChange'], [data-cy]",
            timeout=8000
        )
    except Exception:
        print("[SCRAPER] Stats selector timeout, doing direct extraction")

    extraction_map = {
        "price": [
            "span[data-cy='token-price']",
            "[class*='tokenPrice']",
        ],
        "mcap": [
            "[data-cy='market-cap']",
            "[class*='marketCap']",
        ],
        "volume_24h": [
            "[data-cy='volume-24h']",
            "[class*='volume']",
        ],
        "liquidity": [
            "[data-cy='liquidity']",
            "[class*='liquidity']",
        ],
    }

    for metric, selectors in extraction_map.items():
        for selector in selectors:
            try:
                element = await page.query_selector(selector)
                if element:
                    text = (await element.inner_text()).strip()
                    if text and text != "-":
                        metrics[metric] = text
                        break
            except Exception:
                continue

    # JavaScript extraction — read labeled stat pairs from Dexscreener DOM
    try:
        js_metrics = await page.evaluate("""
            () => {
                const result = {};

                // Dexscreener renders stats as 2-children elements: [label, value].
                // Walk all elements and match known label keywords.
                const elements = Array.from(document.querySelectorAll('*'));
                for (let i = 0; i < elements.length; i++) {
                    const el = elements[i];
                    if (el.children.length !== 2) continue;

                    const label = (el.children[0].innerText || '').trim().toLowerCase();
                    const value = (el.children[1].innerText || '').trim();
                    if (!label || !value || value === '-') continue;

                    if ((label.includes('mkt cap') || label.includes('market cap') || label === 'mcap') && !result.mcap)
                        result.mcap = value;
                    if ((label.includes('price') || label === 'usd') && !result.price)
                        result.price = value;
                    if (label.includes('volume') && !result.volume)
                        result.volume = value;
                    if (label.includes('liquidity') && !result.liquidity)
                        result.liquidity = value;
                    if ((label.includes('txns') || label.includes('buy') || label.includes('sell')) && !result.txns)
                        result.txns = value;
                }

                // Text-scan fallback for MCAP if pair-matching missed it
                if (!result.mcap) {
                    const body = document.body.innerText || '';
                    const m = body.match(/(?:MKT CAP|MCAP|Market Cap)[\\s\\S]{0,40}?(\\$[\\d,.]+[KMBkmb]?)/i);
                    if (m) result.mcap = m[1];
                }

                return result;
            }
        """)
        if js_metrics:
            metrics.update(js_metrics)
    except Exception as e:
        print(f"[SCRAPER] JS eval error: {e}")

    return metrics


async def _take_chart_screenshot(page: Page) -> str:
    """Screenshot the chart element with multiple selector fallbacks."""
    selectors_to_try = [
        "[class*='tradingview']",
        "[class*='chart-pane']",
        "#tv_chart_container",
        "canvas",
        "[class*='chart']",
    ]

    for sel in selectors_to_try:
        try:
            element = await page.query_selector(sel)
            if element:
                img_bytes = await element.screenshot()
                return base64.b64encode(img_bytes).decode("utf-8")
        except Exception:
            continue

    # Final fallback: viewport crop
    try:
        img_bytes = await page.screenshot(
            clip={"x": 0, "y": 100, "width": 1200, "height": 650}
        )
        return base64.b64encode(img_bytes).decode("utf-8")
    except Exception as e:
        print(f"[SCRAPER] All screenshot attempts failed: {e}")
        return ""
