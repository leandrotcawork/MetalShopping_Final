from __future__ import annotations

import argparse
import csv
import hashlib
import json
import os
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

import psycopg


def _now_utc() -> datetime:
    return datetime.now(timezone.utc)


def _ts_id() -> str:
    return _now_utc().strftime("%Y%m%d_%H%M%S")


def _norm_text(value: Any) -> str:
    return ("" if value is None else str(value)).strip()


def _norm_lower(value: Any) -> str:
    return _norm_text(value).lower()


def _bool_from_legacy(value: Any) -> bool:
    raw = _norm_lower(value)
    if raw in ("1", "true", "t", "yes", "y"):
        return True
    if raw in ("0", "false", "f", "no", "n"):
        return False
    try:
        return bool(int(raw))
    except Exception:
        return bool(raw)


def _sha_prefix(prefix: str, value: str, length: int = 24) -> str:
    digest = hashlib.sha256(_norm_lower(value).encode("utf-8")).hexdigest()[:length]
    return f"{prefix}_{digest}"


def _taxonomy_id(legacy_id: Any) -> Optional[str]:
    if legacy_id is None:
        return None
    token = _norm_text(legacy_id)
    if token == "":
        return None
    return f"tx_{token}"


def _product_id(pn_interno: str) -> str:
    return _sha_prefix("prd", pn_interno)


def _identifier_id(product_id: str, identifier_type: str, identifier_value: str) -> str:
    payload = f"{product_id}|{identifier_type}|{identifier_value}"
    return _sha_prefix("pid", payload)


def _normalize_name(name: str) -> str:
    return " ".join(_norm_text(name).split()).lower()


def _set_tenant(cur: psycopg.Cursor, tenant_id: str) -> None:
    cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))


@dataclass
class LegacyTaxonomyNode:
    legacy_id: str
    name: str
    name_norm: str
    code: str
    parent_id: Optional[str]
    level: int
    path_text: str
    is_active: bool


@dataclass
class LegacyTaxonomyLevel:
    level: int
    label: str
    short_label: str
    is_enabled: bool


@dataclass
class LegacyProduct:
    pn_interno: str
    reference: str
    ean: str
    descricao: str
    marca: str
    taxonomy_node_id: Optional[str]
    tipo_estoque: str
    ativo: bool


def _fetchall(conn: psycopg.Connection, sql: str, params: Iterable[Any] | None = None) -> List[Dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(sql, params or [])
        cols = [c.name for c in cur.description]
        return [dict(zip(cols, row)) for row in cur.fetchall()]


def _fetchall_with_schema(
    conn: psycopg.Connection,
    schema: str,
    sql: str,
    params: Iterable[Any] | None = None,
) -> List[Dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(f"SET search_path TO {schema}, public;")
        cur.execute(sql, params or [])
        cols = [c.name for c in cur.description]
        return [dict(zip(cols, row)) for row in cur.fetchall()]


def _legacy_levels(conn: psycopg.Connection, schema: str, client_id: str) -> List[LegacyTaxonomyLevel]:
    sql = """
        SELECT level, label, COALESCE(short_label, '') AS short_label, is_enabled
        FROM taxonomy_level_defs
        WHERE client_id = %s
        ORDER BY level ASC
    """
    rows = _fetchall_with_schema(conn, schema, sql, [client_id])
    levels: List[LegacyTaxonomyLevel] = []
    for row in rows:
        levels.append(
            LegacyTaxonomyLevel(
                level=int(row.get("level") or 0),
                label=_norm_text(row.get("label")),
                short_label=_norm_text(row.get("short_label")),
                is_enabled=bool(row.get("is_enabled", True)),
            )
        )
    return levels


def _legacy_nodes(conn: psycopg.Connection, schema: str, client_id: str) -> List[LegacyTaxonomyNode]:
    sql = """
        SELECT id, name, COALESCE(name_norm, '') AS name_norm, COALESCE(code, '') AS code,
               parent_id, level, COALESCE(path_text, '') AS path_text, is_active
        FROM taxonomy_nodes
        WHERE client_id = %s
        ORDER BY level ASC, parent_id NULLS FIRST, id ASC
    """
    rows = _fetchall_with_schema(conn, schema, sql, [client_id])
    nodes: List[LegacyTaxonomyNode] = []
    for row in rows:
        name = _norm_text(row.get("name"))
        name_norm = _norm_text(row.get("name_norm")) or _normalize_name(name)
        nodes.append(
            LegacyTaxonomyNode(
                legacy_id=_norm_text(row.get("id")),
                name=name,
                name_norm=name_norm,
                code=_norm_text(row.get("code")),
                parent_id=_norm_text(row.get("parent_id")) or None,
                level=int(row.get("level") or 0),
                path_text=_norm_text(row.get("path_text")),
                is_active=bool(row.get("is_active", True)),
            )
        )
    return nodes


def _legacy_products(conn: psycopg.Connection, schema: str) -> List[LegacyProduct]:
    sql = """
        SELECT pn_interno, COALESCE(reference, '') AS reference, COALESCE(ean, '') AS ean,
               COALESCE(descricao, '') AS descricao, COALESCE(marca, '') AS marca,
               taxonomy_node_id, COALESCE(tipo_estoque, '') AS tipo_estoque,
               ativo
        FROM products
        ORDER BY pn_interno ASC
    """
    rows = _fetchall_with_schema(conn, schema, sql)
    products: List[LegacyProduct] = []
    for row in rows:
        products.append(
            LegacyProduct(
                pn_interno=_norm_text(row.get("pn_interno")),
                reference=_norm_text(row.get("reference")),
                ean=_norm_text(row.get("ean")),
                descricao=_norm_text(row.get("descricao")),
                marca=_norm_text(row.get("marca")),
                taxonomy_node_id=_norm_text(row.get("taxonomy_node_id")) or None,
                tipo_estoque=_norm_text(row.get("tipo_estoque")),
                ativo=_bool_from_legacy(row.get("ativo")),
            )
        )
    return products


def _ensure_output_dir(root: Path) -> Path:
    root.mkdir(parents=True, exist_ok=True)
    return root


def _write_csv(path: Path, rows: List[Dict[str, Any]]) -> None:
    if not rows:
        return
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=list(rows[0].keys()))
        writer.writeheader()
        writer.writerows(rows)


def _write_json(path: Path, payload: Any) -> None:
    with path.open("w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=2)


def _collect_legacy_conflicts(products: List[LegacyProduct]) -> Dict[str, List[Dict[str, Any]]]:
    def group_by(field: str) -> List[Dict[str, Any]]:
        bucket: Dict[str, List[str]] = {}
        for p in products:
            value = _norm_text(getattr(p, field))
            if value == "":
                continue
            bucket.setdefault(value, []).append(p.pn_interno)
        conflicts: List[Dict[str, Any]] = []
        for value, ids in bucket.items():
            if len(ids) > 1:
                conflicts.append({"identifier_value": value, "pn_interno_list": ids})
        return conflicts

    return {
        "duplicate_reference": group_by("reference"),
        "duplicate_ean": group_by("ean"),
    }


def _reset_destination(conn: psycopg.Connection, tenant_id: str) -> None:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        cur.execute("DELETE FROM catalog_products")
        cur.execute("DELETE FROM catalog_taxonomy_nodes")
        cur.execute("DELETE FROM catalog_taxonomy_level_defs")
        cur.execute("COMMIT")


def _upsert_taxonomy_levels(
    conn: psycopg.Connection,
    tenant_id: str,
    levels: List[LegacyTaxonomyLevel],
) -> int:
    if not levels:
        return 0
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        for level in levels:
            cur.execute(
                """
                INSERT INTO catalog_taxonomy_level_defs (
                  tenant_id, level, label, short_label, is_enabled, created_at, updated_at
                ) VALUES (
                  current_tenant_id(), %s, %s, %s, %s, NOW(), NOW()
                )
                ON CONFLICT (tenant_id, level) DO UPDATE
                SET label = EXCLUDED.label,
                    short_label = EXCLUDED.short_label,
                    is_enabled = EXCLUDED.is_enabled,
                    updated_at = NOW()
                """,
                (level.level, level.label, level.short_label or None, level.is_enabled),
            )
        cur.execute("COMMIT")
    return len(levels)


def _upsert_taxonomy_nodes(
    conn: psycopg.Connection,
    tenant_id: str,
    nodes: List[LegacyTaxonomyNode],
) -> int:
    if not nodes:
        return 0
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        for node in nodes:
            taxonomy_node_id = _taxonomy_id(node.legacy_id)
            parent_id = _taxonomy_id(node.parent_id) if node.parent_id else None
            cur.execute(
                """
                INSERT INTO catalog_taxonomy_nodes (
                  taxonomy_node_id, tenant_id, name, name_norm, code,
                  parent_taxonomy_node_id, level, path, is_active, created_at, updated_at
                ) VALUES (
                  %s, current_tenant_id(), %s, %s, %s,
                  %s, %s, %s, %s, NOW(), NOW()
                )
                ON CONFLICT (taxonomy_node_id) DO UPDATE
                SET tenant_id = EXCLUDED.tenant_id,
                    name = EXCLUDED.name,
                    name_norm = EXCLUDED.name_norm,
                    code = EXCLUDED.code,
                    parent_taxonomy_node_id = EXCLUDED.parent_taxonomy_node_id,
                    level = EXCLUDED.level,
                    path = EXCLUDED.path,
                    is_active = EXCLUDED.is_active,
                    updated_at = NOW()
                """,
                (
                    taxonomy_node_id,
                    node.name,
                    node.name_norm or _normalize_name(node.name),
                    node.code or None,
                    parent_id,
                    node.level,
                    node.path_text or None,
                    node.is_active,
                ),
            )
        cur.execute("COMMIT")
    return len(nodes)


def _product_name_from_legacy(product: LegacyProduct) -> str:
    for candidate in (product.descricao, product.reference, product.ean, product.pn_interno):
        if _norm_text(candidate):
            return _norm_text(candidate)
    return product.pn_interno


def _insert_identifier(
    cur: psycopg.Cursor,
    *,
    product_id: str,
    identifier_type: str,
    identifier_value: str,
    source_system: str,
    is_primary: bool,
) -> Optional[str]:
    identifier_value = _norm_text(identifier_value)
    if identifier_value == "":
        return "empty"

    cur.execute(
        """
        SELECT product_id
        FROM catalog_product_identifiers
        WHERE tenant_id = current_tenant_id()
          AND identifier_type = %s
          AND identifier_value = %s
        LIMIT 1
        """,
        (identifier_type, identifier_value),
    )
    row = cur.fetchone()
    if row and row[0] != product_id:
        return str(row[0])

    identifier_id = _identifier_id(product_id, identifier_type, identifier_value)
    cur.execute(
        """
        INSERT INTO catalog_product_identifiers (
          product_identifier_id, product_id, tenant_id,
          identifier_type, identifier_value, source_system, is_primary,
          created_at, updated_at
        ) VALUES (
          %s, %s, current_tenant_id(),
          %s, %s, %s, %s,
          NOW(), NOW()
        )
        ON CONFLICT (tenant_id, identifier_type, identifier_value) DO UPDATE
        SET product_id = EXCLUDED.product_id,
            source_system = COALESCE(EXCLUDED.source_system, catalog_product_identifiers.source_system),
            is_primary = EXCLUDED.is_primary,
            updated_at = NOW()
        """,
        (
            identifier_id,
            product_id,
            identifier_type,
            identifier_value,
            source_system,
            is_primary,
        ),
    )
    return None


def _upsert_products(
    conn: psycopg.Connection,
    tenant_id: str,
    products: List[LegacyProduct],
    taxonomy_index: Dict[str, str],
    *,
    commit_every: int,
) -> Tuple[int, List[Dict[str, Any]], List[Dict[str, Any]]]:
    conflicts: List[Dict[str, Any]] = []
    missing_taxonomy: List[Dict[str, Any]] = []
    inserted = 0

    cur = conn.cursor()
    cur.execute("BEGIN")
    _set_tenant(cur, tenant_id)

    for idx, product in enumerate(products, start=1):
        pn = _norm_text(product.pn_interno)
        if pn == "":
            continue

        product_id = _product_id(pn)
        name = _product_name_from_legacy(product)
        description = _norm_text(product.descricao) or None
        brand_name = _norm_text(product.marca) or None
        stock_profile = _norm_text(product.tipo_estoque) or None
        status = "active" if product.ativo else "inactive"

        tax_id = None
        if product.taxonomy_node_id:
            tax_id = taxonomy_index.get(_norm_text(product.taxonomy_node_id))
            if not tax_id:
                missing_taxonomy.append(
                    {
                        "pn_interno": pn,
                        "taxonomy_node_id": _norm_text(product.taxonomy_node_id),
                        "reason": "missing_taxonomy_node",
                    }
                )
        else:
            missing_taxonomy.append({"pn_interno": pn, "taxonomy_node_id": "", "reason": "missing_taxonomy_node_id"})

        cur.execute(
            """
            INSERT INTO catalog_products (
              product_id, tenant_id, sku, name, description,
              brand_name, stock_profile_code, primary_taxonomy_node_id,
              status, created_at, updated_at
            ) VALUES (
              %s, current_tenant_id(), %s, %s, %s,
              %s, %s, %s,
              %s, NOW(), NOW()
            )
            ON CONFLICT (product_id) DO UPDATE
            SET sku = EXCLUDED.sku,
                name = EXCLUDED.name,
                description = COALESCE(EXCLUDED.description, catalog_products.description),
                brand_name = COALESCE(EXCLUDED.brand_name, catalog_products.brand_name),
                stock_profile_code = COALESCE(EXCLUDED.stock_profile_code, catalog_products.stock_profile_code),
                primary_taxonomy_node_id = COALESCE(EXCLUDED.primary_taxonomy_node_id, catalog_products.primary_taxonomy_node_id),
                status = EXCLUDED.status,
                updated_at = NOW()
            """,
            (
                product_id,
                pn,
                name,
                description,
                brand_name,
                stock_profile,
                tax_id,
                status,
            ),
        )

        source_system = "legacy_import_v1"
        _insert_identifier(
            cur,
            product_id=product_id,
            identifier_type="pn_interno",
            identifier_value=pn,
            source_system=source_system,
            is_primary=True,
        )

        for id_type, id_value, primary in (
            ("reference", product.reference, False),
            ("ean", product.ean, False),
        ):
            existing = _insert_identifier(
                cur,
                product_id=product_id,
                identifier_type=id_type,
                identifier_value=id_value,
                source_system=source_system,
                is_primary=primary,
            )
            if existing not in (None, "empty"):
                conflicts.append(
                    {
                        "pn_interno": pn,
                        "identifier_type": id_type,
                        "identifier_value": _norm_text(id_value),
                        "existing_product_id": existing,
                        "incoming_product_id": product_id,
                    }
                )

        inserted += 1

        if commit_every > 0 and idx % commit_every == 0:
            cur.execute("COMMIT")
            cur.execute("BEGIN")
            _set_tenant(cur, tenant_id)

    cur.execute("COMMIT")
    cur.close()
    return inserted, conflicts, missing_taxonomy


def main() -> None:
    parser = argparse.ArgumentParser(description="Import legacy catalog master data into the new catalog tables.")
    parser.add_argument("--database-url", default=os.getenv("MS_DATABASE_URL", "").strip())
    parser.add_argument("--legacy-database-url", default=os.getenv("MS_LEGACY_DATABASE_URL", "").strip())
    parser.add_argument("--tenant-id", default=os.getenv("MS_TENANT_ID", "").strip() or "tenant_default")
    parser.add_argument("--legacy-schema", default=os.getenv("MS_LEGACY_SCHEMA", "").strip() or "metalshopping")
    parser.add_argument("--legacy-client-id", default=os.getenv("MS_LEGACY_CLIENT_ID", "").strip() or "default")
    parser.add_argument("--reset", action="store_true", help="Delete destination catalog data for the tenant before import.")
    parser.add_argument("--dry-run", action="store_true", help="Run pre-checks and reports only.")
    parser.add_argument("--commit-every", type=int, default=500)
    args = parser.parse_args()

    database_url = _norm_text(args.database_url)
    legacy_url = _norm_text(args.legacy_database_url)
    if database_url == "":
        raise SystemExit("Missing --database-url (or MS_DATABASE_URL).")
    if legacy_url == "":
        raise SystemExit("Missing --legacy-database-url (or MS_LEGACY_DATABASE_URL).")

    tenant_id = _norm_text(args.tenant_id) or "tenant_default"
    legacy_schema = _norm_text(args.legacy_schema) or "metalshopping"
    legacy_client_id = _norm_text(args.legacy_client_id) or "default"

    output_root = _ensure_output_dir(Path(".tmp") / "import_catalog_reports" / _ts_id())

    with psycopg.connect(legacy_url) as legacy_conn:
        levels = _legacy_levels(legacy_conn, legacy_schema, legacy_client_id)
        nodes = _legacy_nodes(legacy_conn, legacy_schema, legacy_client_id)
        products = _legacy_products(legacy_conn, legacy_schema)

    duplicate_report = _collect_legacy_conflicts(products)
    _write_json(output_root / "legacy_identifier_conflicts.json", duplicate_report)
    _write_csv(output_root / "legacy_duplicate_reference.csv", duplicate_report["duplicate_reference"])
    _write_csv(output_root / "legacy_duplicate_ean.csv", duplicate_report["duplicate_ean"])

    taxonomy_index = {node.legacy_id: _taxonomy_id(node.legacy_id) for node in nodes}

    if args.dry_run:
        summary = {
            "tenant_id": tenant_id,
            "legacy_schema": legacy_schema,
            "legacy_client_id": legacy_client_id,
            "levels": len(levels),
            "nodes": len(nodes),
            "products": len(products),
            "duplicate_reference": len(duplicate_report["duplicate_reference"]),
            "duplicate_ean": len(duplicate_report["duplicate_ean"]),
            "output_dir": str(output_root),
        }
        _write_json(output_root / "dry_run_summary.json", summary)
        print(json.dumps(summary, indent=2))
        return

    with psycopg.connect(database_url) as dest_conn:
        if args.reset:
            _reset_destination(dest_conn, tenant_id)

        level_count = _upsert_taxonomy_levels(dest_conn, tenant_id, levels)
        node_count = _upsert_taxonomy_nodes(dest_conn, tenant_id, nodes)

        inserted, id_conflicts, missing_tax = _upsert_products(
            dest_conn,
            tenant_id,
            products,
            taxonomy_index,
            commit_every=int(args.commit_every),
        )

    _write_json(output_root / "identifier_conflicts.json", id_conflicts)
    _write_csv(output_root / "identifier_conflicts.csv", id_conflicts)
    _write_json(output_root / "missing_taxonomy.json", missing_tax)
    _write_csv(output_root / "missing_taxonomy.csv", missing_tax)

    summary = {
        "tenant_id": tenant_id,
        "legacy_schema": legacy_schema,
        "legacy_client_id": legacy_client_id,
        "levels": level_count,
        "nodes": node_count,
        "products_imported": inserted,
        "identifier_conflicts": len(id_conflicts),
        "missing_taxonomy_rows": len(missing_tax),
        "output_dir": str(output_root),
    }
    _write_json(output_root / "import_summary.json", summary)
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
