"""
AgenTrading — One-command launcher
Jalankan: python start.py
Hentikan: Ctrl+C (semua proses akan ikut mati)
"""

import subprocess
import sys
import os
import threading
import signal
from pathlib import Path

ROOT = Path(__file__).parent.resolve()

SERVICES = [
    {
        "name": "BACKEND ",
        "color": "\033[94m",   # biru
        "cmd": ["go", "run", "main.go"],
        "cwd": ROOT / "backend",
    },
    {
        "name": "SCRAPER ",
        "color": "\033[92m",   # hijau
        "cmd": [sys.executable, "main.py"],
        "cwd": ROOT / "scraper",
    },
    {
        "name": "FRONTEND",
        "color": "\033[95m",   # magenta
        "cmd": [sys.executable, "-m", "streamlit", "run", "app.py"],
        "cwd": ROOT / "frontend_streamlit",
    },
]

RESET = "\033[0m"
BOLD  = "\033[1m"

processes = []


def stream_output(proc, label, color):
    """Baca output dari subprocess dan print dengan label berwarna."""
    for line in iter(proc.stdout.readline, b""):
        text = line.decode("utf-8", errors="replace").rstrip()
        if text:
            print(f"{color}{BOLD}[{label}]{RESET} {text}")
    for line in iter(proc.stderr.readline, b""):
        text = line.decode("utf-8", errors="replace").rstrip()
        if text:
            print(f"{color}{BOLD}[{label}]{RESET} \033[91m{text}{RESET}")


def shutdown(sig=None, frame=None):
    print(f"\n{BOLD}🛑 Menghentikan semua service...{RESET}")
    for p in processes:
        try:
            p.terminate()
        except Exception:
            pass
    for p in processes:
        try:
            p.wait(timeout=5)
        except Exception:
            p.kill()
    print(f"{BOLD}✅ Semua service dihentikan.{RESET}")
    sys.exit(0)


def main():
    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    print(f"\n{BOLD}🚀 AgenTrading Copilot — Starting all services...{RESET}\n")

    for svc in SERVICES:
        print(f"  {svc['color']}▶ {svc['name']}{RESET}  →  {' '.join(str(c) for c in svc['cmd'])}")

    print()

    threads = []
    for svc in SERVICES:
        try:
            proc = subprocess.Popen(
                svc["cmd"],
                cwd=svc["cwd"],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=os.environ.copy(),
            )
            processes.append(proc)

            t = threading.Thread(
                target=stream_output,
                args=(proc, svc["name"], svc["color"]),
                daemon=True,
            )
            t.start()
            threads.append(t)

        except FileNotFoundError as e:
            print(f"\033[91m[ERROR] Gagal start {svc['name']}: {e}{RESET}")
            shutdown()

    print(f"{BOLD}✅ Semua service jalan! Buka: http://localhost:8501{RESET}")
    print(f"{BOLD}   Tekan Ctrl+C untuk menghentikan semua.{RESET}\n")

    # Tunggu sampai salah satu proses mati
    while True:
        for i, proc in enumerate(processes):
            if proc.poll() is not None:
                name = SERVICES[i]["name"]
                print(f"\033[91m[WARN] {name} berhenti tiba-tiba (exit code {proc.returncode}). Menghentikan semua...{RESET}")
                shutdown()
        threading.Event().wait(2)


if __name__ == "__main__":
    main()
