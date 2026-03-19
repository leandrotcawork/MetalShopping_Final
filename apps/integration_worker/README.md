# integration_worker

Worker dedicado a conectores externos, crawlers, import/export e normalizacao de dados de integracao.

## Shopping Price worker (Level 1 scaffold)

Arquivo:

- `shopping_price_worker.py`

Regra de operacao:

- worker escreve no Postgres (`shopping_price_runs`, `shopping_price_run_items`, `shopping_price_latest_snapshot`)
- `server_core` apenas le e expoe API
- sem chamada HTTP do worker para o backend

Variaveis de ambiente:

- `MS_DATABASE_URL`: DSN do Postgres
- `MS_SHOPPING_INPUT_PATH`: caminho de arquivo JSON de entrada

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

Execucao:

```powershell
python -m pip install -r apps\integration_worker\requirements.txt
$env:MS_DATABASE_URL="postgres://..."
$env:MS_SHOPPING_INPUT_PATH="C:\temp\shopping_input.json"
python apps\integration_worker\shopping_price_worker.py
```
