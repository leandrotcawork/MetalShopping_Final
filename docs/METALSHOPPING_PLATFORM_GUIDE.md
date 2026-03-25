# MetalShopping Platform Guide

Estado observado em: 2026-03-25

Este documento consolida o que o MetalShopping e hoje, quais modulos existem de fato no repositorio, como esses modulos funcionam juntos e quais sao as projecoes futuras mais consistentes com o codigo, os documentos em `docs/` e o arquivo `MetalShopping_Project_Bible.docx`.

## 1. O que e o MetalShopping

MetalShopping nao e um e-commerce tradicional. O produto foi desenhado como uma plataforma empresarial de inteligencia operacional para empresas de materiais de acabamento.

Na pratica, o sistema quer concentrar em um unico produto:

- catalogo canonico de produtos
- precificacao
- estoque
- monitoramento de mercado
- shopping e coleta de sinais de fornecedores
- analytics operacional e estrategico
- CRM e automacoes futuras
- governanca de regras, limites e feature flags

A identidade congelada do projeto hoje e:

- monorepo
- server-first
- `apps/server_core` como nucleo canonico
- workers especializados fora do core
- Postgres como estado transacional canonico
- Go no core
- Python em workers de integracao e compute
- frontend thin-client consumindo SDK gerado

Em termos de filosofia de entrega, o projeto segue explicitamente:

`make it work -> make it clean -> make it fast`

## 2. Tese arquitetural do produto

O MetalShopping foi organizado para evitar dois problemas classicos:

- espalhar regra de negocio entre frontend, backend e automacoes
- deixar workers virarem donos do estado do produto

Por isso, a arquitetura atual parte destes principios:

- `contracts/` e a fonte de verdade para APIs, eventos e governanca
- `server_core` implementa contratos publicos, auth, tenancy, governanca e mutacoes canonicas
- workers processam integracao, scraping, compute e entrega assincrona
- o frontend nao fala com endpoints via `fetch()` manual; ele consome `@metalshopping/sdk-runtime`
- eventos relevantes saem do core via outbox e alimentam workers ou read models futuros

O fluxo padrao do produto hoje e este:

```text
contracts/api/openapi/*.yaml
    -> packages/generated/sdk_ts/generated/*
    -> packages/platform-sdk
    -> apps/web

contracts/events/v1/*.json
    -> outbox no server_core
    -> workers / consumidores assincronos

contracts/governance/*
    -> bootstrap/seeds/governance/*
    -> runtime em apps/server_core/internal/platform/governance/*
```

## 3. Estrutura atual do repositorio

O repositorio esta dividido em blocos com responsabilidades bem definidas:

| Bloco | Papel atual |
|---|---|
| `apps/server_core` | Backend principal, auth, tenancy, governanca, APIs, estado transacional e outbox |
| `apps/web` | Cliente web thin, com shell, rotas e composicao de paginas |
| `apps/integration_worker` | Worker Python de integracao e scraping, hoje com runtime real para Shopping Price |
| `apps/analytics_worker` | Reserva para compute analitico pesado e projecoes futuras |
| `apps/automation_worker` | Reserva para campanhas, gatilhos e automacoes futuras |
| `apps/notifications_worker` | Reserva para entrega por email, SMS, WhatsApp e webhooks |
| `apps/admin_console` | Console administrativo planejado para governanca e operacao |
| `apps/desktop` | Cliente thin planejado para futuro |
| `contracts/` | Contratos HTTP, eventos versionados e artefatos de governanca |
| `packages/ui` | Componentes visuais compartilhados |
| `packages/platform-sdk` | Runtime autoral que encapsula os clientes gerados |
| `packages/feature-*` | Adaptadores, view models e composicao por feature |
| `packages/generated` | SDKs e tipos gerados; nunca editados manualmente |
| `docs/` | SoT, ADRs, planos, regras operacionais e visoes de dominio |

## 4. Modulos que existem hoje no core

Os diretorios de modulo em `apps/server_core/internal/modules/` representam a topologia alvo do produto. Em 2026-03-25, o estado real nao e homogeneo: alguns modulos ja tem slice funcional, outros ainda sao placeholders arquiteturais.

### 4.1 Modulos com slice funcional observada

| Modulo | Estado atual | Papel real hoje |
|---|---|---|
| `iam` | Implementado | Papeis e permissoes internas do MetalShopping |
| `catalog` | Implementado | Produto canonico, taxonomia, identificadores e portfolio base |
| `pricing` | Implementado | Preco interno, custo de reposicao, custo medio e historico de preco |
| `inventory` | Implementado | Posicao de estoque atual e historico de posicoes |
| `home` | Implementado | KPIs resumidos de entrada da plataforma |
| `shopping` | Implementado | Orquestracao de requests de coleta, leitura de runs e exportacoes |
| `suppliers` | Implementado | Diretorio de fornecedores e manifestos de drivers |
| `analytics_serving` | Implementado | Superficie de leitura de Analytics Home |

### 4.2 Modulos reservados na arquitetura, mas sem slice funcional relevante observada

| Modulo | Leitura correta do estado atual |
|---|---|
| `tenant_admin` | Reservado para administracao de tenant |
| `sales` | Reservado para dominio comercial futuro |
| `customers` | Base para CRM futuro |
| `procurement` | Planejado, com fronteira documental congelada, sem runtime material ainda |
| `market_intelligence` | Planejado para evolucao futura |
| `crm` | Planejado para camada futura |
| `automation` | Planejado para automacoes governadas |
| `integrations_control` | Planejado para operacao de conectores |
| `alerts` | Planejado para alertas operacionais e estrategicos |

## 5. Plataforma base dentro do `server_core`

Antes dos modulos de negocio, o produto ja possui uma camada de plataforma bastante definida.

### 5.1 Auth e sessao web

O login atual usa Keycloak como IdP inicial, mas a fronteira de sessao pertence ao `server_core`.

Fluxo real:

1. O browser inicia `GET /api/v1/auth/session/login`.
2. O `server_core` redireciona para o provedor OIDC.
3. O callback volta para `GET /api/v1/auth/session/callback`.
4. O backend valida estado, troca o codigo por token, cria sessao em Postgres e seta cookie `HttpOnly`.
5. O frontend sobe via `GET /api/v1/auth/session/me`.
6. Mutacoes de sessao usam CSRF cookie + header.

Isso evita que o browser vire dono de token, permissao ou tenancy.

### 5.2 Tenancy runtime

A plataforma e multi-tenant desde a fundacao. O padrao atual combina:

- `tenant_id` no modelo de dados
- runtime tenant no contexto
- sessao Postgres tenant-aware
- `current_tenant_id()` como filtro em tabelas tenant-scoped
- RLS como base do isolamento inicial

Em termos de design, isso prepara o sistema para crescimento sem duplicar a aplicacao por cliente.

### 5.3 Runtime governance

A governanca do produto nao foi deixada como configuracao espalhada em codigo.

Hoje existem:

- feature flags em `contracts/governance/feature_flags`
- thresholds em `contracts/governance/thresholds`
- policies em `contracts/governance/policies`
- seeds iniciais em `bootstrap/seeds/governance`
- resolvers em `internal/platform/governance/*`

Essa camada ja interfere em runtime real. Exemplos:

- habilitacao de criacao de produto
- controles de timeout de sessao
- politica de override manual de preco

### 5.4 Outbox e mensageria

Mutacoes relevantes nao deveriam depender de chamadas sincronas a workers. O projeto ja implementa a fundacao de outbox em `internal/platform/messaging/outbox`.

Na pratica, o fluxo esperado e:

- o modulo grava no banco
- no mesmo contexto transacional registra o evento no outbox
- o dispatcher publica ou entrega para consumidores assincronos

Isso ja sustenta eventos como `catalog_product_created`, `pricing_price_set`, `inventory_position_updated` e `shopping_run_requested`.

## 6. Modulos de negocio atuais e como funcionam

### 6.1 IAM

O modulo `iam` e o responsavel pelas permissoes internas do produto. Ele nao substitui o IdP; ele complementa a identidade externa com semantica interna de autorizacao.

O que faz hoje:

- faz upsert de atribuicao de papeis por usuario
- responde a checks de permissao usados por outros modulos
- governa o acesso a operacoes de catalogo, pricing, inventory e administracao

Papel arquitetural:

- Keycloak autentica
- `iam` decide o que o usuario pode fazer dentro do MetalShopping

### 6.2 Catalog

`catalog` e o primeiro modulo canonico forte do produto. Ele e dono da identidade do produto.

O que o modulo ja possui:

- CRUD inicial de produtos
- taxonomia
- identificadores do produto
- descricao
- status do produto
- portfolio base para a superficie `Products`

Responsabilidade real:

- `catalog` nao e so cadastro; ele define a base sobre a qual `pricing`, `inventory`, `shopping`, `analytics` e `procurement` trabalham

Importante:

- criacao de produto ja e governada por feature flag e thresholds
- o modulo ja publica evento via outbox

### 6.3 Pricing

`pricing` e dono da semantica de preco interno.

Capacidades atuais:

- registrar preco de produto
- listar historico de precos
- ler preco atual
- armazenar custo de reposicao
- armazenar custo medio quando disponivel
- aplicar guardas de governanca para override manual
- evitar historico artificial em reruns sem mudanca real

Papel do modulo:

- transformar preco em dado operacional governado, e nao em campo espalhado pelo sistema

### 6.4 Inventory

`inventory` e dono da posicao de estoque viva.

Capacidades atuais:

- registrar posicao de estoque por produto
- listar historico de posicoes
- ler posicao atual
- armazenar `on_hand_quantity`
- armazenar `last_purchase_at`
- armazenar `last_sale_at`
- manter status da posicao

Papel do modulo:

- ser a verdade canonica de estoque operacional

Importante:

- o plano oficial impede que semantica de compras, lead time ou ERP invada `inventory`

### 6.5 Home

`home` e a primeira superficie de entrada do produto.

Hoje o endpoint de resumo entrega:

- quantidade total de produtos
- quantidade de produtos ativos
- quantidade de produtos com preco
- quantidade de produtos com estoque rastreado
- timestamp da ultima atualizacao

Papel do modulo:

- ser um sumario operacional inicial da plataforma
- comprovar a tese `contract -> backend -> sdk -> tela`

### 6.6 Shopping

`shopping` e um dos modulos mais importantes hoje porque ele materializa o padrao server_core + worker.

O que ele faz atualmente:

- expor bootstrap operacional
- expor resumo de runs
- criar `run requests`
- listar runs
- detalhar um run
- listar itens por run
- resumir status por item e por fornecedor
- ler snapshot mais recente por produto
- gerenciar sinais de fornecedor
- listar candidatos de URL manual
- exportar resultados em XLSX

Fluxo funcional atual:

1. O usuario inicia uma coleta pelo frontend.
2. O `server_core` valida auth e tenant.
3. O modulo `shopping` cria um `run_request`.
4. O core registra evento `shopping.run_requested` no outbox.
5. O `integration_worker` processa em modo fila ou modo evento.
6. O worker escreve resultados no Postgres.
7. O `server_core` volta a servir summary, runs, items, latest snapshots e exportacoes.

Esse fluxo e importante porque representa o modelo desejado para integracoes futuras:

- Python faz scraping e processamento de campo
- Go continua sendo a fronteira canonica de API
- o estado fica em Postgres

### 6.7 Suppliers

`suppliers` hoje nao e so um cadastro passivo. Ele funciona como camada de governanca operacional para o runtime de fornecedores.

O que ja existe:

- diretorio de fornecedores
- enable/disable de fornecedor
- manifestos versionados de driver
- validacao de manifestos por familia
- ativacao de manifesto

Familias de runtime observadas:

- `http`
- `playwright`

Estrategias ja previstas/registradas:

- `http.mock.v1`
- `http.vtex_persisted_query.v1`
- `http.html_search.v1`
- `http.leroy_search_sellers.v1`
- `http.html_dom_first_card.v1`
- `playwright.mock.v1`
- `playwright.pdp_first.v1`

Na pratica, `suppliers` virou a camada que permite escalar o Shopping sem hardcode de fornecedor espalhado no worker.

### 6.8 Analytics Serving

`analytics_serving` e a primeira interface backend do dominio analitico.

Hoje ele entrega:

- endpoint de `Analytics Home`
- snapshot de leitura
- blocos estruturados como KPIs, acoes do dia, alertas, distribuicao de portfolio e timeline

Papel atual:

- servir uma superficie de leitura tenant-safe
- alimentar o frontend de analytics sem colocar logica analitica pesada no browser

Papel futuro:

- virar a camada de serving do modulo de inteligencia comercial

## 7. Workers atuais e como entram no fluxo

### 7.1 Integration Worker

E o worker Python mais concreto hoje no repositorio.

Funciona assim:

- pode operar em `queue` mode ou `event` mode
- processa `shopping_price_run_requests`
- usa estrategias `http` e `playwright`
- decide o termo de busca via politica de lookup
- executa a estrategia
- grava observacoes em tabelas de shopping no Postgres

Ponto arquitetural chave:

- ele nao chama endpoint HTTP do core para concluir o trabalho
- ele escreve no banco; o core le e expoe

Esse e o padrao oficial para integracao assincrona nivel 1.

### 7.2 Analytics Worker

Hoje esta mais como boundary reservada do que como slice operacional ativa, mas sua missao esta muito clara nos docs:

- scoring
- projecoes
- explainability
- simulacoes
- processamento pesado
- atualizacao de read models analiticos

### 7.3 Automation Worker

Planejado para:

- gatilhos
- campanhas
- acoes governadas
- follow-ups assincronos

### 7.4 Notifications Worker

Planejado para:

- email
- SMS
- WhatsApp
- webhooks

O dominio de alerta continua no core; este worker so entrega.

## 8. Frontend atual e boundary de cliente fino

O frontend foi explicitamente desenhado para nao virar uma segunda aplicacao de negocio.

### 8.1 `apps/web`

E o thin client principal hoje.

Rotas observadas:

- `/home`
- `/products`
- `/shopping`
- `/analytics`
- `/analytics/products/:pn/*`
- `/login`

Ele e dono de:

- shell
- roteamento
- providers
- composicao de paginas

Ele nao deve ser dono de:

- contratos canonicos
- regras de negocio
- chamadas HTTP manuais

### 8.2 `packages/platform-sdk`

Esse pacote e central para a arquitetura web atual.

Ele:

- encapsula clientes gerados
- padroniza headers, trace id, CSRF e credentials
- mapeia erros HTTP para um runtime consistente
- expande a superficie usada pelo frontend com metodos mais ergonomicos

Sem ele, o frontend dependeria diretamente dos gerados; com ele, existe uma camada autoral estavel e controlada.

### 8.3 `packages/ui`

Centraliza componentes compartilhados ja reutilizados entre superficies, como:

- `AppFrame`
- `Button`
- `Checkbox`
- `FilterDropdown`
- `MetricCard`
- `MetricChip`
- `SelectMenu`
- `SortHeaderButton`
- `StatusBanner`
- `StatusPill`
- `SurfaceCard`

### 8.4 Feature packages

`feature-auth-session`

- login
- bootstrap de sessao
- rota autenticada
- telas de redirect/auth bootstrap

`feature-products`

- adaptadores e widgets da superficie Products
- composicao do portfolio

`feature-analytics`

- adaptadores e DTOs de compatibilidade com o frontend legado
- superficies atuais de Analytics
- workspace de produto
- trilha de migracao visual com foco em paridade

## 9. Contratos atuais e estado de integracao

### 9.1 OpenAPI surfaces observadas

Hoje o repositorio ja possui contratos para:

- `analytics`
- `auth_session`
- `catalog`
- `home`
- `iam`
- `inventory`
- `pricing`
- `products`
- `shopping`
- `suppliers`

### 9.2 Eventos observados

Ja existem eventos versionados para:

- `catalog_product_created`
- `iam_role_assigned`
- `inventory_position_updated`
- `pricing_price_set`
- `shopping_run_requested`

### 9.3 Governanca observada

Ja existem artefatos de:

- feature flags
- thresholds
- policies

Ou seja: o projeto ja passou da fase de "pasta pronta" e entrou em uma fase onde contratos, persistencia, frontend e fluxo assincrono realmente conversam.

## 10. Como o sistema funciona ponta a ponta hoje

### 10.1 Fluxo HTTP sincrono padrao

1. O usuario entra pelo `apps/web`.
2. A sessao e autenticada via `auth/session`.
3. O frontend chama `sdk.*` via `@metalshopping/sdk-runtime`.
4. O `server_core` autentica o principal e resolve o tenant.
5. O modulo de negocio executa validacoes e regras.
6. O acesso a dados acontece em Postgres tenant-aware.
7. O backend responde DTO alinhado ao contrato.

### 10.2 Fluxo de mutacao com evento

1. O cliente chama uma mutacao do core.
2. O modulo grava estado transacional.
3. O evento correspondente entra no outbox.
4. O consumidor assincrono processa depois.
5. O sistema evita acoplamento sincrono entre request e worker.

### 10.3 Fluxo de read surface composta

O melhor exemplo hoje e `Products`.

Ela existe porque o backend consolida:

- identidade do produto via `catalog`
- preco atual via `pricing`
- posicao atual via `inventory`

Assim, o frontend nao precisa costurar tres modulos manualmente.

### 10.4 Fluxo de Shopping Price

Esse e o fluxo mais representativo do modelo alvo:

```text
apps/web
  -> sdk-runtime.shopping.createRunRequest()
  -> server_core/shopping
  -> outbox: shopping.run_requested
  -> integration_worker
  -> Postgres: runs, items, snapshots, signals
  -> server_core/shopping read APIs
  -> apps/web exibe progresso e resultado
```

## 11. Limites de ownership que o projeto ja congelou

Um ponto forte do MetalShopping e que varios limites de ownership ja foram congelados cedo. Isso reduz o risco de drift semantico.

Exemplos importantes:

- `catalog` e dono da identidade do produto
- `pricing` e dono do preco interno e semantica de custo
- `inventory` e dono da posicao de estoque
- `procurement` sera dono de semantica de reposicao e lead time, nao `inventory`
- `analytics` sera dono de metricas derivadas, score, recomendacao e explainability, nao de escrita canonica
- workers nao viram donos da verdade de negocio

## 12. Onde o projeto esta hoje

Cruzar o codigo com o SoT mostra um estado bem claro:

- a fundacao de auth, tenancy, governance e outbox esta montada
- `catalog`, `pricing` e `inventory` ja existem como modulos reais
- `Products` ja e uma superficie real usando backend e SDK
- `Home` ja tem endpoint e pagina
- `Shopping` ja tem backend funcional e worker real
- `Analytics` ja tem serving inicial e forte migracao de frontend em andamento
- `CRM`, `procurement`, `alerts`, `automation` e partes mais fortes de AI ainda estao na fronteira futura

Em resumo: o produto ja saiu da fase conceitual. Ele esta em fundacao implementada com slices reais de produto.

## 13. Projecoes futuras de funcionamento

As projecoes mais consistentes nao sao chute; elas aparecem repetidamente no codigo, nos ADRs, no SoT e no Project Bible.

### 13.1 Horizonte imediato

O horizonte imediato do produto e fechar a Camada 1 operacional.

Isso significa:

- consolidar `Home`
- fechar `Shopping` com paridade operacional maior
- continuar a subida de `Analytics` sobre endpoints reais
- abrir `CRM v1`

Leitura correta: o projeto quer primeiro colocar de pe os modulos que o time interno usa no dia a dia.

### 13.2 Horizonte de dominio

Depois da camada operacional inicial, a tendencia mais clara do repo e esta:

- `procurement` nasce em cima de entradas publicadas e nao em leitura direta de ERP
- `analytics` cresce de dashboard para motor de decisao
- `suppliers` evolui de diretorio/manifesto para malha governada de conectores
- `shopping` deixa de ser uma tela e vira workflow operacional completo
- `products` segue como read surface consolidada sobre dominios canonicos

### 13.3 Horizonte analitico

O repositorio ja descreve `analytics` como um modulo de inteligencia em oito camadas:

- metricas
- regras
- calculos
- classificacoes
- recomendacoes
- campanhas
- alertas
- AI

Se essa visao for seguida, o futuro do produto e:

- mostrar o que esta piorando e por que
- sugerir acao por SKU, marca e taxonomia
- priorizar filas de trabalho
- gerar explainability rastreavel
- produzir sugestoes de compra e de preco com evidencia

### 13.4 Horizonte de automacao e AI

O Project Bible aponta uma linha bem objetiva:

- campanhas automaticas
- agentes de marketplace
- pesquisa de mercado com AI
- perfil preditivo de cliente
- recomendacoes de compra
- alertas inteligentes

Pelo desenho atual do repo, isso provavelmente acontecera assim:

- `analytics_worker` produz score e recomendacao
- `automation_worker` transforma recomendacao em acao governada
- `notifications_worker` entrega alerta ou campanha
- `server_core` continua como fronteira oficial de leitura, aprovacao e auditoria

### 13.5 Horizonte multi-cliente

Hoje o foco e o time interno, mas a arquitetura ja foi desenhada para futuro multi-tenant mais amplo.

Os principais sinais disso sao:

- tenancy desde a base
- governanca runtime
- contratos versionados
- thin clients
- separacao entre identidade externa e IAM interno

Isso prepara o produto para sair de um software de operacao interna e virar plataforma comercializavel sem reescrever os fundamentos.

## 14. Leitura executiva final

O MetalShopping hoje ja e uma plataforma empresarial em construcao avancada, nao um prototipo desorganizado.

Sua forma atual pode ser resumida assim:

- o core em Go ja controla auth, tenancy, governanca, contratos e dados canonicos
- o web ja funciona como thin client apoiado em SDK gerado
- o Shopping ja prova o modelo `Go + Python + Postgres + outbox`
- os dominios basicos `catalog`, `pricing` e `inventory` ja sustentam a primeira superficie forte do produto
- a proxima grande evolucao e transformar `Analytics`, `Shopping`, `CRM` e depois `procurement` em motores operacionais completos

Se o plano continuar coerente com o que ja foi congelado, o resultado nao sera apenas um sistema de cadastro e dashboard. Sera uma plataforma de inteligencia comercial, execucao operacional e automacao governada para o varejo de materiais de acabamento.
