from __future__ import annotations

import json
import re
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

from ..headers import build_runtime_headers
from ..models import RuntimeObservation, SupplierRuntimeConfig, SupplierSignal
from ..shared.parsing import b64json, decode_price_text, extract_by_paths, safe_float, safe_int, safe_str


def _retry_http_statuses(config: SupplierRuntimeConfig) -> set[int]:
    raw = config.config_json.get("retryHttpStatuses")
    default_statuses = {408, 425, 429, 500, 502, 503, 504}
    if not isinstance(raw, list):
        return default_statuses
    parsed: set[int] = set()
    for item in raw:
        try:
            status = int(item)
        except (TypeError, ValueError):
            continue
        if 100 <= status <= 599:
            parsed.add(status)
    return parsed if len(parsed) > 0 else default_statuses


def _strategy_for_config(config: SupplierRuntimeConfig) -> str:
    return safe_str(config.config_json.get("strategy"), "").lower()


def _vtex_build_params(
    *,
    term: str,
    operation_name: str,
    sha256_hash: str,
    skus_filter: str,
    to_n: int,
    include_variant: bool,
) -> dict[str, str]:
    variables: dict[str, Any] = {
        "hideUnavailableItems": False,
        "skusFilter": skus_filter,
        "simulationBehavior": "default",
        "installmentCriteria": "MAX_WITHOUT_INTEREST",
        "productOriginVtex": False,
        "map": "ft",
        "query": term,
        "orderBy": "OrderByScoreDESC",
        "from": 0,
        "to": int(to_n),
        "selectedFacets": [{"key": "ft", "value": term}],
        "fullText": term,
        "facetsBehavior": "Static",
        "categoryTreeBehavior": "default",
        "withFacets": False,
    }
    if include_variant:
        variables["variant"] = "null-null"

    extensions = {
        "persistedQuery": {
            "version": 1,
            "sha256Hash": sha256_hash,
            "sender": "vtex.store-resources@0.x",
            "provider": "vtex.search-graphql@0.x",
        },
        "variables": b64json(variables),
    }

    return {
        "workspace": "master",
        "maxAge": "short",
        "appsEtag": "remove",
        "domain": "store",
        "locale": "pt-BR",
        "operationName": operation_name,
        "variables": "{}",
        "extensions": json.dumps(extensions, separators=(",", ":")),
    }


def _vtex_pick_seller(
    sellers: list[dict[str, Any]],
    *,
    preferred_name: str,
) -> dict[str, Any] | None:
    preferred = preferred_name.strip().lower()
    if preferred:
        for seller in sellers:
            if str(seller.get("sellerName") or "").strip().lower() == preferred:
                return seller
    for seller in sellers:
        if seller.get("sellerDefault") is True:
            return seller
    return sellers[0] if sellers else None


def _vtex_offer_available(offer: dict[str, Any]) -> bool:
    is_available = offer.get("IsAvailable")
    if isinstance(is_available, bool) and not is_available:
        return False
    qty = offer.get("AvailableQuantity")
    try:
        if qty is not None and float(qty) <= 0:
            return False
    except Exception:
        pass
    availability = str(offer.get("Availability") or offer.get("availability") or "").strip().lower()
    if availability and ("unavailable" in availability or "outofstock" in availability):
        return False
    return True


def _vtex_extract_price(payload: dict[str, Any], ean: str, *, config: SupplierRuntimeConfig) -> tuple[float | None, str]:
    data = payload.get("data") if isinstance(payload.get("data"), dict) else {}
    product_search = data.get("productSearch") if isinstance(data.get("productSearch"), dict) else {}
    products = product_search.get("products") if isinstance(product_search.get("products"), list) else []

    preferred_seller = safe_str(config.config_json.get("preferredSellerName"), "")
    require_available = bool(config.config_json.get("requireAvailableOffer", False))
    allow_fallback = bool(config.config_json.get("allowFallbackFirstProduct", False))
    token = str(ean or "").strip()

    def pick_from_product(p: dict[str, Any]) -> tuple[float | None, str]:
        items = p.get("items") if isinstance(p.get("items"), list) else []
        for item in items:
            if not isinstance(item, dict):
                continue
            if token and str(item.get("ean") or "").strip() != token:
                continue
            sellers = item.get("sellers") if isinstance(item.get("sellers"), list) else []
            sellers = [s for s in sellers if isinstance(s, dict)]
            chosen = _vtex_pick_seller(sellers, preferred_name=preferred_seller)
            offer = chosen.get("commertialOffer") if isinstance(chosen, dict) and isinstance(chosen.get("commertialOffer"), dict) else {}
            if require_available and not _vtex_offer_available(offer):
                return (None, "ean_exact_but_unavailable")
            price = offer.get("Price")
            parsed = safe_float(price, -1.0)
            if parsed <= 0:
                return (None, "ean_exact_but_no_price")
            return (parsed, "ean_exact_match")
        return (None, "ean_not_found_in_items")

    for product in products:
        if not isinstance(product, dict):
            continue
        price, note = pick_from_product(product)
        if price is not None:
            return price, note

    if allow_fallback and products and isinstance(products[0], dict):
        first = products[0]
        items = first.get("items") if isinstance(first.get("items"), list) else []
        item0 = items[0] if items and isinstance(items[0], dict) else {}
        sellers = item0.get("sellers") if isinstance(item0.get("sellers"), list) else []
        sellers = [s for s in sellers if isinstance(s, dict)]
        chosen = _vtex_pick_seller(sellers, preferred_name=preferred_seller)
        offer = chosen.get("commertialOffer") if isinstance(chosen, dict) and isinstance(chosen.get("commertialOffer"), dict) else {}
        price = safe_float(offer.get("Price"), -1.0)
        if price > 0:
            return price, "fallback_first_product"

    return None, "no_price_found"


def _render_runtime_url(
    config: SupplierRuntimeConfig,
    product_id: str,
    signal: SupplierSignal | None,
) -> str | None:
    if signal is not None and signal.product_url:
        return signal.product_url
    endpoint_template = safe_str(config.config_json.get("endpointTemplate"), "")
    if endpoint_template != "":
        encoded_product = urllib.parse.quote(product_id, safe="")
        encoded_supplier = urllib.parse.quote(config.supplier_code, safe="")
        return endpoint_template.replace("{product_id}", encoded_product).replace("{supplier_code}", encoded_supplier)
    base_url = safe_str(config.config_json.get("baseUrl"), "")
    if base_url != "":
        separator = "&" if "?" in base_url else "?"
        encoded_product = urllib.parse.quote(product_id, safe="")
        encoded_supplier = urllib.parse.quote(config.supplier_code, safe="")
        return f"{base_url}{separator}product_id={encoded_product}&supplier_code={encoded_supplier}"
    return None


def _build_vtex_url(
    config: SupplierRuntimeConfig,
    term: str,
    signal: SupplierSignal | None,
) -> str | None:
    if signal is not None and signal.product_url:
        return signal.product_url

    base_url = safe_str(config.config_json.get("baseUrl"), "")
    operation_name = safe_str(config.config_json.get("operationName"), "")
    sha256_hash = safe_str(config.config_json.get("sha256Hash"), "")
    if base_url == "" or operation_name == "" or sha256_hash == "":
        return None

    params = _vtex_build_params(
        term=str(term or "").strip(),
        operation_name=operation_name,
        sha256_hash=sha256_hash,
        skus_filter=safe_str(config.config_json.get("skusFilter"), "FIRST_AVAILABLE"),
        to_n=safe_int(config.config_json.get("toN"), 11, 1, 60),
        include_variant=bool(config.config_json.get("includeVariant", False)),
    )
    return f"{base_url}?{urllib.parse.urlencode(params)}"


def _build_html_search_url(
    config: SupplierRuntimeConfig,
    term: str,
    signal: SupplierSignal | None,
) -> str | None:
    if signal is not None and signal.product_url:
        return signal.product_url
    template = safe_str(config.config_json.get("searchUrlTemplate"), "")
    if template == "":
        return None
    lookup_mode = signal.lookup_mode if signal is not None else "REFERENCE"
    return (
        template.replace("{product_id}", urllib.parse.quote(term, safe=""))
        .replace("{term}", urllib.parse.quote(term, safe=""))
        .replace("{supplier_code}", urllib.parse.quote(config.supplier_code, safe=""))
        .replace("{lookup_mode}", urllib.parse.quote(lookup_mode, safe=""))
    )


def _execute_http_vtex_strategy(
    *,
    config: SupplierRuntimeConfig,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
    timeout_seconds: int,
    max_retries: int,
    seller_default: str,
) -> RuntimeObservation:
    url = _build_vtex_url(config, lookup_term, signal)
    if url is None:
        return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP_VTEX", None, "vtex_url_missing", _strategy_for_config(config), lookup_term)

    headers = build_runtime_headers(config)
    headers.setdefault("Accept", "application/json")
    headers.setdefault("Content-Type", "application/json")
    last_error = "vtex_runtime_failed"
    strategy = _strategy_for_config(config)
    retry_statuses = _retry_http_statuses(config)

    for attempt in range(1, max_retries + 1):
        try:
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=timeout_seconds) as response:
                raw_body = response.read().decode("utf-8", errors="replace")
                loaded = json.loads(raw_body)
                body_json = loaded if isinstance(loaded, dict) else {"data": loaded}
                extracted_price, note = _vtex_extract_price(body_json, lookup_term, config=config)
                if extracted_price is None:
                    status = "NOT_FOUND" if note in {"ean_not_found_in_items", "ean_exact_but_no_price", "ean_exact_but_unavailable", "no_price_found"} else "ERROR"
                    return RuntimeObservation(status, safe_float(base_price, base_price), seller_default, "HTTP_VTEX", int(response.status), note, strategy, lookup_term)
                return RuntimeObservation(
                    "OK",
                    safe_float(extracted_price, base_price),
                    safe_str(config.config_json.get("sellerName"), seller_default),
                    "HTTP_VTEX",
                    int(response.status),
                    f"http_vtex_runtime attempt={attempt}",
                    strategy,
                    lookup_term,
                )
        except urllib.error.HTTPError as exc:
            if exc.code not in retry_statuses:
                return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_VTEX", int(exc.code), f"http_status_{exc.code}_not_retryable", strategy, lookup_term)
            last_error = f"http_status_{exc.code}"
        except Exception as exc:
            last_error = f"http_error:{exc}"
        time.sleep(0.1)
    return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_VTEX", None, last_error[:280], strategy, lookup_term)


def _execute_http_html_search_strategy(
    *,
    config: SupplierRuntimeConfig,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
    timeout_seconds: int,
    max_retries: int,
    seller_default: str,
) -> RuntimeObservation:
    url = _build_html_search_url(config, lookup_term, signal)
    if url is None:
        return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP_HTML", None, "html_search_url_missing", _strategy_for_config(config), lookup_term)

    headers = build_runtime_headers(config)
    headers.setdefault("Accept", "text/html,application/json")
    last_error = "html_runtime_failed"
    price_regex = safe_str(config.config_json.get("priceRegex"), r"(\d{1,3}(?:\.\d{3})*,\d{2}|\d+(?:\.\d{2})?)")
    seller_regex = safe_str(config.config_json.get("sellerRegex"), "")
    strategy = _strategy_for_config(config)
    retry_statuses = _retry_http_statuses(config)

    for attempt in range(1, max_retries + 1):
        try:
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=timeout_seconds) as response:
                raw_body = response.read().decode("utf-8", errors="replace")
                parsed_json: dict[str, Any] | None = None
                if raw_body.lstrip().startswith("{"):
                    loaded = json.loads(raw_body)
                    if isinstance(loaded, dict):
                        parsed_json = loaded

                if parsed_json is not None:
                    observed_price, seller_name, channel = extract_by_paths(
                        body_json=parsed_json,
                        config=config,
                        base_price=base_price,
                        seller_default=seller_default,
                        channel_default="HTTP_HTML",
                    )
                    return RuntimeObservation("OK", observed_price, seller_name, channel, int(response.status), f"http_html_json attempt={attempt}", strategy, lookup_term)

                price_match = re.search(price_regex, raw_body, flags=re.IGNORECASE | re.MULTILINE)
                if price_match is None:
                    return RuntimeObservation("NOT_FOUND", safe_float(base_price, base_price), seller_default, "HTTP_HTML", int(response.status), "price_not_found_in_html", strategy, lookup_term)
                observed_price = decode_price_text(price_match.group(1), base_price)

                seller_name = seller_default
                if seller_regex != "":
                    seller_match = re.search(seller_regex, raw_body, flags=re.IGNORECASE | re.MULTILINE)
                    if seller_match is not None and seller_match.lastindex:
                        seller_name = safe_str(seller_match.group(1), seller_default)

                return RuntimeObservation("OK", observed_price, seller_name, "HTTP_HTML", int(response.status), f"http_html_runtime attempt={attempt}", strategy, lookup_term)
        except urllib.error.HTTPError as exc:
            if exc.code not in retry_statuses:
                return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_HTML", int(exc.code), f"http_status_{exc.code}_not_retryable", strategy, lookup_term)
            last_error = f"http_status_{exc.code}"
        except Exception as exc:
            last_error = f"http_error:{exc}"
        time.sleep(0.1)
    return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_HTML", None, last_error[:280], strategy, lookup_term)


def execute_http_runtime(
    *,
    config: SupplierRuntimeConfig,
    strategy: str,
    product_id: str,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
) -> RuntimeObservation:
    timeout_seconds = safe_int(config.config_json.get("timeoutSeconds"), 5, 1, 60)
    max_retries = safe_int(config.config_json.get("maxRetries"), 2, 1, 8)
    seller_default = safe_str(config.config_json.get("sellerName"), config.supplier_code.lower())

    if strategy == "http.mock.v1":
        url = _render_runtime_url(config, product_id, signal)
        if url is None or not url.startswith("mock://"):
            return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP", None, "http.mock.v1 requires mock:// endpointTemplate", strategy, lookup_term)
        multiplier = safe_float(config.config_json.get("mockMultiplier"), 1.03)
        observed_price = safe_float(base_price * multiplier, base_price)
        return RuntimeObservation(
            "OK",
            observed_price,
            seller_default,
            "HTTP_MOCK",
            200,
            f"mock_runtime strategy={strategy} timeout={timeout_seconds}s retries={max_retries}",
            strategy,
            lookup_term,
        )

    if strategy == "http.vtex_persisted_query.v1":
        return _execute_http_vtex_strategy(
            config=config,
            lookup_term=lookup_term,
            base_price=base_price,
            signal=signal,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            seller_default=seller_default,
        )

    if strategy == "http.html_search.v1":
        return _execute_http_html_search_strategy(
            config=config,
            lookup_term=lookup_term,
            base_price=base_price,
            signal=signal,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            seller_default=seller_default,
        )

    return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP", None, f"unsupported_http_strategy:{strategy}", strategy, lookup_term)
