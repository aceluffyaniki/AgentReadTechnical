"""
explorer_client.py — Fetches contract source code from chain explorers.
Supports: BscScan (BSC), Basescan (Base), Helius/Solscan (Solana).
"""

import os
import httpx
import asyncio
from typing import Optional, Tuple
from chain_config import EXPLORER_API_URLS


async def fetch_evm_source(
    token_address: str,
    explorer_type: str,  # "bscscan" | "basescan"
) -> Tuple[str, str]:
    """Fetch Solidity source from BscScan or Basescan."""
    api_key = ""
    if explorer_type == "bscscan":
        api_key = os.getenv("BSCSCAN_API_KEY", "")
    elif explorer_type == "basescan":
        api_key = os.getenv("BASESCAN_API_KEY", "")

    base_url = EXPLORER_API_URLS.get(explorer_type, "")
    if not base_url:
        return "", "unknown"

    params = {
        "module": "contract",
        "action": "getsourcecode",
        "address": token_address,
        "apikey": api_key,
    }

    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            resp = await client.get(base_url, params=params)
            data = resp.json()

        if data.get("status") == "1" and data.get("result"):
            result = data["result"][0]
            source = result.get("SourceCode", "")
            if source and source != "":
                return source, "solidity"
    except Exception as e:
        print(f"[EXPLORER] EVM source fetch error ({explorer_type}): {e}")

    return "", "unknown"


async def fetch_solana_info(token_address: str) -> Tuple[str, str]:
    """
    For Solana, we don't have Solidity. Instead, fetch on-chain token info
    (mint authority, freeze authority, metadata) from Helius or Solscan.
    Returns a structured text summary for Claude to audit.
    """
    helius_key = os.getenv("HELIUS_API_KEY", "")

    summary_lines = []

    # Try Helius token metadata
    if helius_key:
        try:
            url = f"{EXPLORER_API_URLS['helius']}/token-metadata?api-key={helius_key}"
            payload = {"mintAccounts": [token_address], "includeOffChain": True, "disableCache": False}
            async with httpx.AsyncClient(timeout=15.0) as client:
                resp = await client.post(url, json=payload)
                data = resp.json()

            if data and len(data) > 0:
                meta = data[0]
                on_chain = meta.get("onChainMetadata", {}).get("metadata", {}).get("data", {})
                account = meta.get("onChainAccountInfo", {}).get("accountInfo", {}).get("data", {})

                summary_lines.append(f"Token Name: {on_chain.get('name', 'Unknown')}")
                summary_lines.append(f"Symbol: {on_chain.get('symbol', 'Unknown')}")
                summary_lines.append(f"Mint Authority: {account.get('parsed', {}).get('info', {}).get('mintAuthority', 'None/Renounced')}")
                summary_lines.append(f"Freeze Authority: {account.get('parsed', {}).get('info', {}).get('freezeAuthority', 'None/Renounced')}")
                summary_lines.append(f"Decimals: {account.get('parsed', {}).get('info', {}).get('decimals', 'Unknown')}")
                summary_lines.append(f"Supply: {account.get('parsed', {}).get('info', {}).get('supply', 'Unknown')}")
        except Exception as e:
            print(f"[EXPLORER] Helius error: {e}")

    # Fallback to Solscan
    if not summary_lines:
        try:
            url = f"{EXPLORER_API_URLS['solscan']}/token/meta?tokenAddress={token_address}"
            async with httpx.AsyncClient(timeout=15.0) as client:
                resp = await client.get(url)
                data = resp.json()

            if data:
                summary_lines.append(f"Token Name: {data.get('name', 'Unknown')}")
                summary_lines.append(f"Symbol: {data.get('symbol', 'Unknown')}")
                summary_lines.append(f"Mint Authority: {data.get('mintAuthority', 'Unknown')}")
                summary_lines.append(f"Freeze Authority: {data.get('freezeAuthority', 'Unknown')}")
                summary_lines.append(f"Supply: {data.get('supply', 'Unknown')}")
        except Exception as e:
            print(f"[EXPLORER] Solscan error: {e}")

    if not summary_lines:
        summary_lines.append(f"Token Address: {token_address}")
        summary_lines.append("Source: explorers tidak merespons, analisis berdasarkan chart saja")

    return "\n".join(summary_lines), "rust"


async def get_contract_source(
    token_address: str,
    chain: str,
    explorer_type: str,
) -> Tuple[str, str]:
    """
    Main entry point. Returns (source_code_text, source_type).
    source_type: "solidity" | "rust" | "unknown"
    """
    if chain == "solana":
        return await fetch_solana_info(token_address)
    else:
        return await fetch_evm_source(token_address, explorer_type)
