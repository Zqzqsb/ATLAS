#!/usr/bin/env python3
"""Probe LLM chat and embedding endpoints defined in llm_config.json."""

from __future__ import annotations

import argparse
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path

PLACEHOLDER_RE = re.compile(
    r"your-deepseek-api-key|your-qwen-api-key|YOUR_TOKEN_HERE|your-embedding-api-key",
    re.I,
)

GREEN = "\033[1;32m"
RED = "\033[1;31m"
YELLOW = "\033[1;33m"
CYAN = "\033[1;36m"
NC = "\033[0m"


def is_placeholder(value: str) -> bool:
    if not value or not value.strip():
        return True
    return bool(PLACEHOLDER_RE.search(value))


def load_multimodal_flag() -> bool:
    for path in (
        Path("backend/server/configs/lakebase.yaml"),
        Path("lakebase.yaml"),
    ):
        if not path.is_file():
            continue
        match = re.search(r"multimodal:\s*(true|false)", path.read_text(), re.I)
        if match:
            return match.group(1).lower() == "true"
    return False


def post_json(url: str, token: str, payload: dict, timeout: float) -> tuple[int, str]:
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read(512).decode("utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        body = exc.read(512).decode("utf-8", errors="replace")
        return exc.code, body
    except urllib.error.URLError as exc:
        return 0, str(exc.reason)


def summarize_error(status: int, body: str) -> str:
    if status == 0:
        return body
    try:
        parsed = json.loads(body)
        err = parsed.get("error")
        if isinstance(err, dict):
            msg = err.get("message") or err.get("code") or body
            return str(msg)[:120]
    except json.JSONDecodeError:
        pass
    snippet = body.replace("\n", " ").strip()
    if status:
        return f"HTTP {status}: {snippet[:120]}"
    return snippet[:120]


def probe_chat(name: str, cfg: dict, timeout: float) -> tuple[bool, str]:
    token = cfg.get("token", "")
    model = cfg.get("model_name", "")
    base_url = cfg.get("base_url", "").rstrip("/")
    if is_placeholder(token):
        return False, "skipped (placeholder token)"
    if not model or not base_url:
        return False, "missing model_name or base_url"

    url = f"{base_url}/chat/completions"
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": "ping"}],
        "max_tokens": 1,
    }
    status, body = post_json(url, token, payload, timeout)
    if status == 200:
        return True, "ok"
    return False, summarize_error(status, body)


def probe_embedding(cfg: dict, timeout: float, multimodal: bool) -> tuple[bool, str]:
    token = cfg.get("api_key", "")
    model = cfg.get("model", "")
    base_url = cfg.get("base_url", "").rstrip("/")
    if is_placeholder(token):
        return False, "skipped (placeholder api_key)"
    if not model or not base_url:
        return False, "missing model or base_url"

    use_multimodal = multimodal or "vision" in model.lower()
    if use_multimodal:
        url = f"{base_url}/embeddings/multimodal"
        payload = {
            "model": model,
            "encoding_format": "float",
            "input": [{"type": "text", "text": "ping"}],
        }
    else:
        url = f"{base_url}/embeddings"
        payload = {
            "model": model,
            "encoding_format": "float",
            "input": "ping",
        }

    status, body = post_json(url, token, payload, timeout)
    if status == 200:
        return True, "ok"
    return False, summarize_error(status, body)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "config",
        nargs="?",
        default="llm_config.json",
        help="Path to llm_config.json (default: llm_config.json)",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=15.0,
        help="HTTP timeout in seconds (default: 15)",
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Exit with code 1 if any probe fails",
    )
    args = parser.parse_args()

    config_path = Path(args.config)
    if not config_path.is_file():
        print(f"{RED}⚠️  {config_path} not found — skip endpoint probe.{NC}")
        return 0

    try:
        raw = json.loads(config_path.read_text())
    except json.JSONDecodeError as exc:
        print(f"{RED}⚠️  Invalid JSON in {config_path}: {exc}{NC}")
        return 1 if args.strict else 0

    embedding_cfg = raw.pop("_embedding", None)
    multimodal = load_multimodal_flag()
    failures = 0

    print(f"{CYAN}🔎 Probing LLM / embedding endpoints...{NC}")

    for name, cfg in sorted(raw.items()):
        if not isinstance(cfg, dict):
            continue
        ok, detail = probe_chat(name, cfg, args.timeout)
        if ok:
            print(f"  {GREEN}✓{NC} LLM {name} ({cfg.get('model_name', '?')}) — {detail}")
        elif detail.startswith("skipped"):
            print(f"  {YELLOW}○{NC} LLM {name} — {detail}")
        else:
            failures += 1
            print(
                f"  {RED}✗{NC} LLM {name} ({cfg.get('base_url', '?')}) — {detail}"
            )

    if isinstance(embedding_cfg, dict):
        ok, detail = probe_embedding(embedding_cfg, args.timeout, multimodal)
        model = embedding_cfg.get("model", "?")
        if ok:
            endpoint = (
                "embeddings/multimodal"
                if multimodal or "vision" in model.lower()
                else "embeddings"
            )
            print(f"  {GREEN}✓{NC} embedding ({model}, /{endpoint}) — {detail}")
        elif detail.startswith("skipped"):
            print(f"  {YELLOW}○{NC} embedding — {detail}")
        else:
            failures += 1
            print(
                f"  {RED}✗{NC} embedding ({embedding_cfg.get('base_url', '?')}) — {detail}"
            )
    else:
        print(f"  {YELLOW}○{NC} embedding — no _embedding section in config")

    if failures:
        print(
            f"{RED}⚠️  {failures} endpoint probe(s) failed — check tokens, base_url, and network.{NC}"
        )
        return 1 if args.strict else 0

    probed = sum(
        1
        for name, cfg in raw.items()
        if isinstance(cfg, dict) and not is_placeholder(cfg.get("token", ""))
    ) + (
        1
        if isinstance(embedding_cfg, dict)
        and not is_placeholder(embedding_cfg.get("api_key", ""))
        else 0
    )
    if probed:
        print(f"{GREEN}✅ All configured endpoints reachable.{NC}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
