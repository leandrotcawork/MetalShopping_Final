# Apps

`apps/` concentra o core, os workers especializados e os clientes thin do MetalShopping.

- `server_core`: backend canonico da plataforma.
- `analytics_worker`: compute analitico e publicacao de outputs.
- `integration_worker`: conectores, crawlers e normalizacao.
- `automation_worker`: orquestracao, triggers e acoes automaticas.
- `notifications_worker`: entrega assicrona por canal.
- `web`, `desktop`, `admin_console`: clientes finos orientados ao core.

