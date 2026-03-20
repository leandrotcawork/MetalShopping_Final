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


def _leroy_extract_product_id_from_url(url: str) -> str:
    text = safe_str(url, "")
    if text == "":
        return ""
    match = re.search(r"_(\d{5,})(?:[/?#]|$)", text, flags=re.IGNORECASE)
    if match is not None:
        return safe_str(match.group(1), "")
    return ""


def _leroy_extract_product_id_from_html(raw_html: str) -> str:
    patterns = [
        r'"productID"\s*:\s*"(\d{5,})"',
        r'"productId"\s*:\s*"(\d{5,})"',
        r'"sku"\s*:\s*"(\d{5,})"',
        r'"product_id"\s*:\s*"(\d{5,})"',
    ]
    for pattern in patterns:
        match = re.search(pattern, raw_html, flags=re.IGNORECASE | re.MULTILINE)
        if match is not None:
            return safe_str(match.group(1), "")

    fallback = re.search(r"_(\d{5,})(?:[/?#\"'])", raw_html, flags=re.IGNORECASE | re.MULTILINE)
    if fallback is not None:
        return safe_str(fallback.group(1), "")
    return ""


def _build_leroy_search_url(
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


def _build_leroy_sellers_url(config: SupplierRuntimeConfig, product_id: str) -> str | None:
    template = safe_str(config.config_json.get("sellersUrlTemplate"), "")
    if template == "":
        return None
    return template.replace("{product_id}", urllib.parse.quote(product_id, safe=""))


def _leroy_extract_sellers(payload: Any) -> list[dict[str, Any]]:
    if isinstance(payload, list):
        return [item for item in payload if isinstance(item, dict)]
    if not isinstance(payload, dict):
        return []
    if isinstance(payload.get("sellers"), list):
        return [item for item in payload["sellers"] if isinstance(item, dict)]
    data = payload.get("data")
    if isinstance(data, dict) and isinstance(data.get("sellers"), list):
        return [item for item in data["sellers"] if isinstance(item, dict)]
    return []


def _leroy_seller_price(seller: dict[str, Any], base_price: float) -> float:
    direct_keys = ("salePrice", "salesPrice", "price", "value")
    for key in direct_keys:
        parsed = decode_price_text(seller.get(key), -1.0)
        if parsed > 0:
            return parsed

    nested_prices = seller.get("prices")
    if isinstance(nested_prices, dict):
        for key in ("salePrice", "salesPrice", "price", "current"):
            parsed = decode_price_text(nested_prices.get(key), -1.0)
            if parsed > 0:
                return parsed
    return safe_float(base_price, base_price)


def _leroy_seller_name(seller: dict[str, Any], fallback: str) -> str:
    for key in ("sellerName", "name", "seller", "displayName"):
        value = safe_str(seller.get(key), "")
        if value != "":
            return value
    return fallback


def _leroy_pick_seller(
    sellers: list[dict[str, Any]],
    *,
    base_price: float,
    fallback_seller: str,
    pick_strategy: str,
) -> tuple[float, str, str]:
    if len(sellers) == 0:
        return safe_float(base_price, base_price), fallback_seller, "no_sellers_found"

    normalized = safe_str(pick_strategy, "selected").lower()
    normalized = normalized if normalized in {"selected", "min_sale"} else "selected"

    def seller_tuple(item: dict[str, Any]) -> tuple[float, str]:
        return _leroy_seller_price(item, base_price), _leroy_seller_name(item, fallback_seller)

    if normalized == "selected":
        for seller in sellers:
            if bool(seller.get("selected")):
                price, seller_name = seller_tuple(seller)
                return price, seller_name, "selected_seller"

    best_price = float("inf")
    best_name = fallback_seller
    for seller in sellers:
        price, seller_name = seller_tuple(seller)
        if price > 0 and price < best_price:
            best_price = price
            best_name = seller_name

    if best_price != float("inf"):
        return safe_float(best_price, base_price), best_name, "min_sale_seller"

    price0, name0 = seller_tuple(sellers[0])
    return price0, name0, "first_seller_fallback"


def _http_get_with_retries(
    *,
    url: str,
    headers: dict[str, str],
    timeout_seconds: int,
    max_retries: int,
    retry_statuses: set[int],
) -> tuple[str | None, int | None, str | None, str]:
    last_error = "request_failed"
    for _attempt in range(1, max_retries + 1):
        try:
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=timeout_seconds) as response:
                return (
                    response.read().decode("utf-8", errors="replace"),
                    int(response.status),
                    safe_str(response.geturl(), url),
                    "ok",
                )
        except urllib.error.HTTPError as exc:
            if exc.code not in retry_statuses:
                return None, int(exc.code), None, f"http_status_{exc.code}_not_retryable"
            last_error = f"http_status_{exc.code}"
        except Exception as exc:
            last_error = f"http_error:{exc}"
        time.sleep(0.1)
    return None, None, None, last_error


def _execute_http_leroy_search_sellers_strategy(
    *,
    config: SupplierRuntimeConfig,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
    timeout_seconds: int,
    max_retries: int,
    seller_default: str,
) -> RuntimeObservation:
    strategy = _strategy_for_config(config)
    search_url = _build_leroy_search_url(config, lookup_term, signal)
    if search_url is None:
        return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP_LEROY", None, "leroy_search_url_missing", strategy, lookup_term)

    retry_statuses = _retry_http_statuses(config)
    base_headers = build_runtime_headers(config)
    base_headers.setdefault("Accept", "text/html,application/json")

    product_id = _leroy_extract_product_id_from_url(search_url)
    search_status: int | None = None
    final_url = search_url
    search_note = "search_url_from_signal" if signal is not None and signal.product_url else "search_template"

    if product_id == "":
        search_body, status, resolved_url, fetch_note = _http_get_with_retries(
            url=search_url,
            headers=base_headers,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            retry_statuses=retry_statuses,
        )
        search_status = status
        if search_body is None:
            if status is not None:
                return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", status, f"leroy_search_failed:{fetch_note}", strategy, lookup_term)
            return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", None, f"leroy_search_failed:{fetch_note}", strategy, lookup_term)

        final_url = safe_str(resolved_url, search_url)
        product_id = _leroy_extract_product_id_from_url(final_url)
        if product_id != "":
            search_note = "search_final_url_product_id"
        else:
            product_id = _leroy_extract_product_id_from_html(search_body)
            if product_id != "":
                search_note = "search_html_ldjson_product_id"

    if product_id == "":
        return RuntimeObservation("NOT_FOUND", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", search_status, "leroy_product_id_not_found", strategy, lookup_term)

    sellers_url = _build_leroy_sellers_url(config, product_id)
    if sellers_url is None:
        return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", search_status, "leroy_sellers_url_missing", strategy, lookup_term)

    seller_headers = dict(base_headers)
    seller_headers["x-region"] = safe_str(config.config_json.get("region"), "uberlandia")
    seller_headers.setdefault("Accept", "application/json")

    sellers_body, sellers_status, _resolved, sellers_note = _http_get_with_retries(
        url=sellers_url,
        headers=seller_headers,
        timeout_seconds=timeout_seconds,
        max_retries=max_retries,
        retry_statuses=retry_statuses,
    )
    if sellers_body is None:
        if sellers_status is not None:
            return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", sellers_status, f"leroy_sellers_failed:{sellers_note}", strategy, lookup_term)
        return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", None, f"leroy_sellers_failed:{sellers_note}", strategy, lookup_term)

    try:
        payload = json.loads(sellers_body)
    except Exception:
        return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", sellers_status, "leroy_sellers_invalid_json", strategy, lookup_term)

    sellers = _leroy_extract_sellers(payload)
    if len(sellers) == 0:
        return RuntimeObservation("NOT_FOUND", safe_float(base_price, base_price), seller_default, "HTTP_LEROY", sellers_status, "leroy_sellers_empty", strategy, lookup_term)

    pick_strategy = safe_str(config.config_json.get("sellerPickStrategy"), "selected")
    observed_price, seller_name, pick_note = _leroy_pick_seller(
        sellers,
        base_price=base_price,
        fallback_seller=seller_default,
        pick_strategy=pick_strategy,
    )
    note = f"leroy_ok search={search_note} sellers={pick_note} product_id={product_id} final_url={final_url}"
    return RuntimeObservation("OK", observed_price, seller_name, "HTTP_LEROY", sellers_status, note[:280], strategy, lookup_term)


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

    if strategy == "http.leroy_search_sellers.v1":
        return _execute_http_leroy_search_sellers_strategy(
            config=config,
            lookup_term=lookup_term,
            base_price=base_price,
            signal=signal,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            seller_default=seller_default,
        )

    return RuntimeObservation("ERROR", safe_float(base_price, 0), seller_default, "HTTP", None, f"unsupported_http_strategy:{strategy}", strategy, lookup_term)
