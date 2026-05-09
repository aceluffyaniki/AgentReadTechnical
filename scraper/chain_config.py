"""
Chain-specific configuration for DEX scraping.
Each chain has its own DEX URL template and explorer API endpoint.
"""

CHAIN_CONFIG = {
    "solana": {
        "dex": "dexscreener",
        "dex_url_template": "https://dexscreener.com/solana/{pair_or_token}",
        "explorer_type": "helius",    # Uses Helius API for on-chain data
        "chart_screenshot_selector": "[class*='chart-container']",
        "metrics_selectors": {
            "price":     "[data-cy='token-price'], .price-value",
            "mcap":      "[data-cy='market-cap'], .mc-value",
            "volume_24h": "[data-cy='volume'], .vol-value",
            "liquidity": "[data-cy='liquidity'], .liq-value",
            "holders":   ".holders-value",
        },
    },
    "base": {
        "dex": "dexscreener",
        "dex_url_template": "https://dexscreener.com/base/{pair_or_token}",
        "explorer_type": "basescan",
        "chart_screenshot_selector": "[class*='chart-container']",
        "metrics_selectors": {
            "price":     "[data-cy='token-price'], .price-value",
            "mcap":      "[data-cy='market-cap'], .mc-value",
            "volume_24h": "[data-cy='volume'], .vol-value",
            "liquidity": "[data-cy='liquidity'], .liq-value",
        },
    },
    "bsc": {
        "dex": "dexscreener",
        "dex_url_template": "https://dexscreener.com/bsc/{pair_or_token}",
        "explorer_type": "bscscan",
        "chart_screenshot_selector": "[class*='chart-container']",
        "metrics_selectors": {
            "price":     "[data-cy='token-price'], .price-value",
            "mcap":      "[data-cy='market-cap'], .mc-value",
            "volume_24h": "[data-cy='volume'], .vol-value",
            "liquidity": "[data-cy='liquidity'], .liq-value",
        },
    },
}

EXPLORER_API_URLS = {
    "basescan": "https://api.basescan.org/api",
    "bscscan":  "https://api.bscscan.com/api",
    "helius":   "https://api.helius.xyz/v0",
    "solscan":  "https://public-api.solscan.io",
}
