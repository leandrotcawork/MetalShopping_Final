# integration_worker

Worker dedicado a conectores externos, crawlers, import/export e normalizacao de dados de integracao.

## Shopping Price worker (ADR-0018 + ADR-0025)

Arquivo:

- `shopping_price_worker.py`

Regra de operacao:

- queue mode: claim em `shopping_price_run_requests` (`queued -> claimed -> running -> completed/failed`)
- event mode: consome eventos publicados `shopping.run_requested` do outbox e processa o `run_request_id` alvo
- worker escreve no Postgres (`shopping_price_runs`, `shopping_price_run_items`, `shopping_price_latest_snapshot`)
- `server_core` apenas le e expoe API
- sem chamada HTTP do worker para o backend

Variaveis de ambiente:

- `MS_DATABASE_URL`: DSN do Postgres
- `MS_TENANT_ID`: tenant alvo no modo fila (obrigatorio sem `MS_SHOPPING_INPUT_PATH`)
- `MS_SHOPPING_WORKER_MODE`: `queue` (default) ou `event`
- `MS_WORKER_ID`: identificador do worker (opcional)
- `MS_SHOPPING_MAX_QUEUE_CLAIMS`: quantidade maxima de requests por execucao (opcional, default `1`)
- `MS_SHOPPING_XLSX_FALLBACK_LIMIT`: limite de itens no fallback de `xlsx` (opcional, default `50`)
- `MS_SHOPPING_INPUT_PATH`: caminho de arquivo JSON de entrada (modo legado)

Formato minimo do JSON:

```json
{
  "tenant_id": "tenant_default",
  "run_id": "d3f5d9ec-4f7d-4ce1-8788-93f8d227f8f6",
  "run_status": "completed",
  "started_at": "2026-03-19T12:00:00Z",
  "finished_at": "2026-03-19T12:02:00Z",
  "notes": "crawler local",
  "items": [
    {
      "product_id": "00000000-0000-0000-0000-000000000001",
      "seller_name": "Marketplace X",
      "channel": "web",
      "observed_price": 129.9,
      "currency_code": "BRL",
      "observed_at": "2026-03-19T12:01:00Z"
    }
  ]
}
```

Execucao (modo fila ADR-0018):

```powershell
python -m pip install -r apps\integration_worker\requirements.txt
$env:MS_DATABASE_URL="postgres://..."
$env:MS_TENANT_ID="tenant_default"
$env:MS_SHOPPING_MAX_QUEUE_CLAIMS="5"
python apps\integration_worker\shopping_price_worker.py
```

Execucao (modo evento ADR-0025):

```powershell
python -m pip install -r apps\integration_worker\requirements.txt
$env:MS_DATABASE_URL="postgres://..."
$env:MS_SHOPPING_WORKER_MODE="event"
$env:MS_SHOPPING_MAX_QUEUE_CLAIMS="10"
# opcional: filtrar por tenant especifico
$env:MS_TENANT_ID="tenant_default"
python apps\integration_worker\shopping_price_worker.py
```

Execucao (modo legado via JSON):

```powershell
$env:MS_DATABASE_URL="postgres://..."
$env:MS_SHOPPING_INPUT_PATH="C:\temp\shopping_input.json"
python apps\integration_worker\shopping_price_worker.py
```
