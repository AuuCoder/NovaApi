from __future__ import annotations

import asyncio
import base64
import json
import os
import random
import re
import string
import time
import urllib.request
import uuid
from typing import Any

import orjson
from curl_cffi.requests import AsyncSession
from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel, Field


CHAT_URL = "https://grok.com/rest/app-chat/conversations/new"
ASSET_BASE_URL = "https://assets.grok.com/"
BRIDGE_KEY = str(os.environ.get("GROK_WEB_BRIDGE_KEY") or "").strip()
FLARESOLVERR_URL = str(os.environ.get("GROK_FLARESOLVERR_URL") or "").strip().rstrip("/")
USER_AGENT = str(os.environ.get("GROK_WEB_USER_AGENT") or "").strip() or (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
)
MAX_IMAGE_BYTES = 25 * 1024 * 1024
CLEARANCE_TTL_SECONDS = 3600

_clearance_cache: dict[str, tuple[float, dict[str, str], str]] = {}
_clearance_lock = asyncio.Lock()


class GenerateRequest(BaseModel):
    sso_token: str = Field(..., min_length=16)
    prompt: str = Field(..., min_length=1, max_length=20_000)
    n: int = Field(default=1, ge=1, le=4)
    proxy_url: str = ""


app = FastAPI(title="Sub2API Grok Web Bridge", docs_url=None, redoc_url=None)


def _normalize_sso(value: str) -> str:
    token = str(value or "").strip()
    if token.lower().startswith("sso="):
        token = token[4:].split(";", 1)[0].strip()
    return token


def _session_kwargs(proxy_url: str) -> dict[str, Any]:
    proxy = str(proxy_url or "").strip()
    if not proxy:
        return {"impersonate": "chrome"}
    return {
        "impersonate": "chrome",
        "proxies": {"http": proxy, "https": proxy},
    }


def _client_hints(user_agent: str) -> dict[str, str]:
    # Keep the same client-hint version selection as the proven Grok runtime.
    # Its first 2-3 digit match is X11/Windows 10 before the Chrome version.
    match = re.search(r"(\d{2,3})", user_agent)
    if not match:
        return {}
    version = match.group(1)
    platform = "Windows" if "windows" in user_agent.lower() else "Linux"
    arch = "x86" if any(value in user_agent.lower() for value in ("x86_64", "x64", "win64")) else ""
    hints = {
        "Sec-Ch-Ua": f'"Google Chrome";v="{version}", "Chromium";v="{version}", "Not(A:Brand";v="24"',
        "Sec-Ch-Ua-Mobile": "?0",
        "Sec-Ch-Ua-Model": "",
        "Sec-Ch-Ua-Platform": f'"{platform}"',
    }
    if arch:
        hints["Sec-Ch-Ua-Arch"] = arch
        hints["Sec-Ch-Ua-Bitness"] = "64"
    return hints


def _statsig_id() -> str:
    if random.choice((True, False)):
        suffix = "".join(random.choices(string.ascii_lowercase + string.digits, k=5))
        message = f"x1:TypeError: Cannot read properties of null (reading 'children['{suffix}']')"
    else:
        suffix = "".join(random.choices(string.ascii_lowercase, k=10))
        message = f"x1:TypeError: Cannot read properties of undefined (reading '{suffix}')"
    return base64.b64encode(message.encode()).decode()


def _cookie(token: str, clearance_cookies: dict[str, str] | None = None) -> str:
    values = {"sso": token, "sso-rw": token}
    for name, value in (clearance_cookies or {}).items():
        name = str(name or "").strip()
        value = str(value or "").strip()
        if name and value and name.lower() not in {"sso", "sso-rw"}:
            values[name] = value
    return "; ".join(f"{name}={value}" for name, value in values.items())


def _cached_clearance(proxy_url: str) -> tuple[dict[str, str], str]:
    cached = _clearance_cache.get(str(proxy_url or "").strip())
    if not cached or cached[0] <= time.monotonic():
        return {}, USER_AGENT
    return dict(cached[1]), cached[2] or USER_AGENT


async def _refresh_clearance(proxy_url: str) -> tuple[dict[str, str], str]:
    if not FLARESOLVERR_URL:
        raise RuntimeError("Grok Cloudflare clearance runtime is not configured")
    cache_key = str(proxy_url or "").strip()
    async with _clearance_lock:
        cached = _clearance_cache.get(cache_key)
        if cached and cached[0] > time.monotonic():
            return dict(cached[1]), cached[2] or USER_AGENT

        payload: dict[str, Any] = {
            "cmd": "request.get",
            "url": "https://grok.com",
            "maxTimeout": 120_000,
        }
        if cache_key:
            payload["proxy"] = {"url": cache_key}

        def request_clearance() -> dict[str, Any]:
            request = urllib.request.Request(
                f"{FLARESOLVERR_URL}/v1",
                data=json.dumps(payload).encode("utf-8"),
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(request, timeout=130) as response:
                return json.load(response)

        data = await asyncio.to_thread(request_clearance)
        if str(data.get("status") or "").lower() != "ok":
            raise RuntimeError("Grok Cloudflare clearance refresh failed")
        solution = data.get("solution")
        if not isinstance(solution, dict):
            raise RuntimeError("Grok Cloudflare clearance response is invalid")

        cookies: dict[str, str] = {}
        for item in solution.get("cookies") or []:
            if not isinstance(item, dict):
                continue
            domain = str(item.get("domain") or "").lower().lstrip(".")
            name = str(item.get("name") or "").strip()
            value = str(item.get("value") or "").strip()
            if domain and not domain.endswith("grok.com"):
                continue
            if name and value:
                cookies[name] = value
        user_agent = str(solution.get("userAgent") or USER_AGENT).strip() or USER_AGENT
        if not cookies:
            raise RuntimeError("Grok Cloudflare clearance returned no cookies")
        if challenge := cookies.get("x-challenge"):
            # The proven Grok runtime forwards only this challenge cookie.
            # Sending FlareSolverr's analytics/device/signature cookies makes
            # the app-chat endpoint reject the otherwise valid browser proof.
            cookies = {"x-challenge": challenge}

        _clearance_cache[cache_key] = (
            time.monotonic() + CLEARANCE_TTL_SECONDS,
            cookies,
            user_agent,
        )
        return dict(cookies), user_agent


def _chat_headers(token: str, clearance_cookies: dict[str, str], user_agent: str) -> dict[str, str]:
    headers = {
        "Accept": "*/*",
        "Accept-Encoding": "gzip, deflate, br, zstd",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        "Baggage": (
            "sentry-environment=production,"
            "sentry-release=d6add6fb0460641fd482d767a335ef72b9b6abb8,"
            "sentry-public_key=b311e0f2690c81f25e2c4cf6d4f7ce1c"
        ),
        "Content-Type": "application/json",
        "Origin": "https://grok.com",
        "Priority": "u=1, i",
        "Referer": "https://grok.com/",
        "Sec-Fetch-Dest": "empty",
        "Sec-Fetch-Mode": "cors",
        "Sec-Fetch-Site": "same-origin",
        "User-Agent": user_agent,
        "x-statsig-id": _statsig_id(),
        "x-xai-request-id": str(uuid.uuid4()),
    }
    headers.update(_client_hints(user_agent))
    headers["Cookie"] = _cookie(token, clearance_cookies)
    return headers


def _asset_headers(token: str, clearance_cookies: dict[str, str], user_agent: str) -> dict[str, str]:
    headers = {
        "Accept": "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        "Cache-Control": "no-cache",
        "Cookie": _cookie(token, clearance_cookies),
        "Pragma": "no-cache",
        "Priority": "u=0, i",
        "Referer": "https://grok.com/",
        "Sec-Fetch-Dest": "document",
        "Sec-Fetch-Mode": "navigate",
        "Sec-Fetch-Site": "none",
        "Sec-Fetch-User": "?1",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": user_agent,
    }
    headers.update(_client_hints(user_agent))
    return headers


def _payload(prompt: str) -> dict[str, Any]:
    return {
        "collectionIds": [],
        "connectors": [],
        "deviceEnvInfo": {
            "darkModeEnabled": False,
            "devicePixelRatio": 2,
            "screenHeight": 1329,
            "screenWidth": 2056,
            "viewportHeight": 1083,
            "viewportWidth": 2056,
        },
        "disableMemory": True,
        "disableSearch": False,
        "disableSelfHarmShortCircuit": False,
        "disableTextFollowUps": False,
        "enableImageGeneration": True,
        "enableImageStreaming": True,
        "enableSideBySide": True,
        "fileAttachments": [],
        "forceConcise": False,
        "forceSideBySide": False,
        "imageAttachments": [],
        "imageGenerationCount": 2,
        "isAsyncChat": False,
        "message": f"Drawing: {prompt}",
        "modeId": "fast",
        "responseMetadata": {},
        "returnImageBytes": False,
        "returnRawGrokInXaiRequest": False,
        "searchAllConnectors": False,
        "sendFinalMetadata": True,
        "temporary": True,
        "toolOverrides": {
            "gmailSearch": False,
            "googleCalendarSearch": False,
            "outlookSearch": False,
            "outlookCalendarSearch": False,
            "googleDriveSearch": False,
        },
    }


def _image_url_from_frame(data: str) -> str:
    try:
        payload = orjson.loads(data)
    except (orjson.JSONDecodeError, TypeError, ValueError):
        return ""
    error = payload.get("error") if isinstance(payload, dict) else None
    if isinstance(error, dict):
        message = str(error.get("message") or error.get("error") or "Grok stream error")
        raise RuntimeError(message[:300])
    response = ((payload.get("result") or {}).get("response") or {}) if isinstance(payload, dict) else {}
    card = response.get("cardAttachment") if isinstance(response, dict) else None
    if not isinstance(card, dict):
        return ""
    raw = card.get("jsonData")
    try:
        card_data = orjson.loads(raw) if isinstance(raw, (str, bytes)) else raw
    except (orjson.JSONDecodeError, TypeError, ValueError):
        return ""
    chunk = card_data.get("image_chunk") if isinstance(card_data, dict) else None
    if not isinstance(chunk, dict):
        return ""
    try:
        progress = int(chunk.get("progress") or 0)
    except (TypeError, ValueError):
        progress = 0
    image_path = str(chunk.get("imageUrl") or "").strip()
    if progress != 100 or chunk.get("moderated") or not image_path:
        return ""
    if image_path.startswith("http://") or image_path.startswith("https://"):
        return image_path
    return ASSET_BASE_URL + image_path.lstrip("/")


async def _generate_one(token: str, prompt: str, proxy_url: str) -> dict[str, str]:
    clearance_cookies, user_agent = _cached_clearance(proxy_url)
    if FLARESOLVERR_URL and not clearance_cookies:
        clearance_cookies, user_agent = await _refresh_clearance(proxy_url)
    response = None
    session = None
    for attempt in range(2):
        session = AsyncSession(**_session_kwargs(proxy_url))
        response = await session.post(
            CHAT_URL,
            headers=_chat_headers(token, clearance_cookies, user_agent),
            data=orjson.dumps(_payload(prompt)),
            timeout=180,
            stream=True,
        )
        if response.status_code == 403 and attempt == 0:
            await session.close()
            _clearance_cache.pop(str(proxy_url or "").strip(), None)
            clearance_cookies, user_agent = await _refresh_clearance(proxy_url)
            continue
        break

    if session is None or response is None:
        raise RuntimeError("Grok image generation session was not created")
    try:
        if response.status_code != 200:
            body = response.content.decode("utf-8", "replace")[:300]
            raise RuntimeError(f"Grok chat returned HTTP {response.status_code}: {body}")

        image_url = ""
        async for raw_line in response.aiter_lines():
            line = raw_line.decode("utf-8", "replace") if isinstance(raw_line, bytes) else str(raw_line)
            line = line.strip()
            if not line:
                continue
            if line.startswith("data:"):
                line = line[5:].strip()
            if line == "[DONE]":
                break
            if not line.startswith("{"):
                continue
            image_url = _image_url_from_frame(line) or image_url
            if image_url:
                break
        if not image_url:
            raise RuntimeError("Grok image generation returned no image")

        asset = await session.get(
            image_url,
            headers=_asset_headers(token, clearance_cookies, user_agent),
            timeout=120,
        )
        if asset.status_code != 200:
            raise RuntimeError(f"Grok asset download returned HTTP {asset.status_code}")
        raw = bytes(asset.content)
        if not raw or len(raw) > MAX_IMAGE_BYTES:
            raise RuntimeError("Grok asset payload is empty or too large")
        mime_type = str(asset.headers.get("content-type") or "image/jpeg").split(";", 1)[0].strip()
        if not mime_type.startswith("image/"):
            mime_type = "image/jpeg"
        return {
            "mime_type": mime_type,
            "data": base64.b64encode(raw).decode("ascii"),
        }
    finally:
        await session.close()


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/generate")
async def generate(
    body: GenerateRequest,
    x_internal_key: str | None = Header(default=None),
) -> dict[str, Any]:
    if not BRIDGE_KEY or x_internal_key != BRIDGE_KEY:
        raise HTTPException(status_code=401, detail="invalid internal key")
    token = _normalize_sso(body.sso_token)
    if not token:
        raise HTTPException(status_code=400, detail="missing SSO token")
    images: list[dict[str, str]] = []
    try:
        for _ in range(body.n):
            images.append(await _generate_one(token, body.prompt, body.proxy_url))
    except Exception as exc:
        raise HTTPException(status_code=502, detail=str(exc)[:500]) from exc
    return {"images": images}
