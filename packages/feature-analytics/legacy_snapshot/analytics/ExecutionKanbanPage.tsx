import { type MouseEvent, useEffect, useMemo, useState } from "react";

import backlogRaw from "../../../../../../docs/analytics_live/pt-br/todos/analytics/index.md?raw";
import architectureBacklogRaw from "../../../../../../docs/analytics_live/pt-br/todos/arquitetura_reestruturacao/index.md?raw";
import architectureRemediationRaw from "../../../../../../docs/analytics_live/pt-br/todos/arquitetura_reestruturacao/legacy_remediation_extension.md?raw";
import buyingBacklogRaw from "../../../../../../docs/analytics_live/pt-br/todos/buying_engine/index.md?raw";

import styles from "./execution_kanban.module.css";

type KanbanStatus = "doing" | "backlog" | "blocked" | "done";

type KanbanCard = {
  id: string;
  title: string;
  priority: "P0" | "P1" | "P2";
  status: KanbanStatus;
  summary: string;
  outcome: string;
  nextStep: string;
  implementation: string[];
};

type BoardKey = "analytics" | "architecture" | "buying_engine";

type BoardConfig = {
  key: BoardKey;
  label: string;
  title: string;
  idPrefix: string;
  docPath: string;
  docHref: string;
  raw: string;
};

const boardConfigs: BoardConfig[] = [
  {
    key: "analytics",
    label: "Analytics",
    title: "Programa Analytics",
    idPrefix: "TODO-AN-",
    docPath: "docs/analytics_live/pt-br/todos/analytics/index.md",
    docHref: "/docs/analytics_live/pt-br/todos/analytics/index.md",
    raw: backlogRaw
  },
  {
    key: "architecture",
    label: "Arquitetura",
    title: "Programa Arquitetura e Reestruturacao",
    idPrefix: "TODO-AR-",
    docPath: "docs/analytics_live/pt-br/todos/arquitetura_reestruturacao/index.md",
    docHref: "/docs/analytics_live/pt-br/todos/arquitetura_reestruturacao/index.md",
    raw: architectureBacklogRaw + "\n" + architectureRemediationRaw
  },
  {
    key: "buying_engine",
    label: "Buying Engine",
    title: "Programa Buying Engine",
    idPrefix: "TODO-BE-",
    docPath: "docs/analytics_live/pt-br/todos/buying_engine/index.md",
    docHref: "/docs/analytics_live/pt-br/todos/buying_engine/index.md",
    raw: buyingBacklogRaw
  },
];

const statusLabels: Record<KanbanStatus, string> = {
  doing: "Doing",
  backlog: "Backlog",
  blocked: "Blocked",
  done: "Done"
};

function stripInlineMarkdown(value: string): string {
  return value
    .replace(/`([^`]+)`/g, "$1")
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/\*([^*]+)\*/g, "$1");
}

function normalizeText(value: string): string {
  return stripInlineMarkdown(value)
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .join(" ")
    .replace(/\s+/g, " ")
    .trim();
}

function extractSection(block: string, heading: string): string {
  const escapedHeading = heading.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = block.match(new RegExp(`### ${escapedHeading}\\s*\\r?\\n([\\s\\S]*?)(?=\\r?\\n### |$)`));
  return match?.[1]?.trim() ?? "";
}

function extractBullets(section: string, marker: "- " | "- [ ]"): string[] {
  return section
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line.startsWith(marker))
    .map((line) => normalizeText(line.slice(marker.length)));
}

function parseStatus(block: string): KanbanStatus {
  const rawStatus = block.match(/\*\*Status\*\*:\s*([A-Z]+)/)?.[1] ?? "BACKLOG";
  if (rawStatus === "DONE") return "done";
  if (rawStatus === "DOING") return "doing";
  if (rawStatus === "BLOCKED") return "blocked";
  return "backlog";
}

function parsePriority(block: string): KanbanCard["priority"] {
  const rawPriority = block.match(/\*\*Prioridade\*\*:\s*(P[0-2])/)?.[1];
  if (rawPriority === "P0" || rawPriority === "P1" || rawPriority === "P2") return rawPriority;
  return "P1";
}

function parseBacklogCards(raw: string, idPrefix: string): KanbanCard[] {
  const escapedPrefix = idPrefix.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const matches = [...raw.matchAll(new RegExp(`^## (${escapedPrefix}\\d+) - (.+)$`, "gm"))];
  return matches.map((match, index) => {
    const start = match.index ?? 0;
    const end = matches[index + 1]?.index ?? raw.length;
    const block = raw.slice(start, end).trim();
    const scopeItems = extractBullets(extractSection(block, "Escopo"), "- ");
    const acceptanceItems = extractBullets(extractSection(block, "Criterios de aceite"), "- [ ]");
    const summary = normalizeText(extractSection(block, "Contexto"));
    const outcome = normalizeText(extractSection(block, "Objetivo"));

    return {
      id: match[1],
      title: normalizeText(match[2]),
      priority: parsePriority(block),
      status: parseStatus(block),
      summary,
      outcome,
      nextStep: acceptanceItems[0] ?? scopeItems[0] ?? outcome ?? "Revisar backlog oficial",
      implementation: scopeItems.length > 0 ? scopeItems : ["Revisar backlog oficial."]
    };
  });
}

const columns: Array<{ key: KanbanStatus; title: string; subtitle: string }> = [
  { key: "backlog", title: "Backlog", subtitle: "Itens mapeados e ainda nao iniciados" },
  { key: "doing", title: "Doing", subtitle: "Itens em execucao no backlog oficial" },
  { key: "blocked", title: "Blocked", subtitle: "Itens travados por dependencia ou decisao" },
  { key: "done", title: "Done", subtitle: "Itens ja consolidados no programa" }
];

export function ExecutionKanbanPage() {
  const [activeBoard, setActiveBoard] = useState<BoardKey>("architecture");
  const [openCardId, setOpenCardId] = useState<string>("");
  const [docOpenHint, setDocOpenHint] = useState<string>("");
  const board = boardConfigs.find((item) => item.key === activeBoard) ?? boardConfigs[0];
  const cards = useMemo(() => parseBacklogCards(board.raw, board.idPrefix), [board]);
  const cardsCount = cards.length;

  useEffect(() => {
    const nextOpenCardId =
      cards.find((card) => card.status === "doing")?.id ??
      cards.find((card) => card.status === "backlog")?.id ??
      cards[0]?.id ??
      "";
    setOpenCardId(nextOpenCardId);
    setDocOpenHint("");
  }, [cards, activeBoard]);

  function cardsByStatus(status: KanbanStatus): KanbanCard[] {
    return cards.filter((card) => card.status === status);
  }

  function handleOpenBacklogDoc(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    const opened = window.open(board.docHref, "_blank", "noopener,noreferrer");
    if (opened) {
      setDocOpenHint("");
      return;
    }
    void navigator.clipboard?.writeText(board.docPath);
    setDocOpenHint("Nao foi possivel abrir direto no app. Caminho do backlog copiado.");
  }

  return (
    <section className={styles.page}>
      <div className={styles.hero}>
        <div>
          <p className={styles.eyebrow}>Execution Board</p>
          <h1 className={styles.title}>Kanban do {board.title}</h1>
          <p className={styles.subtitle}>
            Quadro leve para visualizar os ANs do backlog oficial por board, sem abrir o documento inteiro.
          </p>
        </div>
        <div className={styles.heroMeta}>
          <div className={styles.metaCard}>
            <span className={styles.metaLabel}>Board ativo</span>
            <div className={styles.boardSwitcher} role="tablist" aria-label="Selecionar board">
              {boardConfigs.map((item) => (
                <button
                  key={item.key}
                  type="button"
                  role="tab"
                  aria-selected={activeBoard === item.key}
                  className={`${styles.boardTab} ${activeBoard === item.key ? styles.boardTabActive : ""}`}
                  onClick={() => setActiveBoard(item.key)}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>
          <div className={styles.metaCard}>
            <span className={styles.metaLabel}>Source of truth</span>
            <strong className={styles.metaValue}>{board.docPath}</strong>
          </div>
          <div className={styles.metaCard}>
            <span className={styles.metaLabel}>Colunas</span>
            <strong className={styles.metaValue}>Backlog / Doing / Blocked / Done</strong>
          </div>
          <div className={styles.metaCard}>
            <span className={styles.metaLabel}>Cards</span>
            <strong className={styles.metaValue}>{cardsCount} ANs visiveis no board</strong>
          </div>
        </div>
      </div>

      <div className={styles.board}>
        {columns.map((column) => (
          <section key={column.key} className={styles.column}>
            <header className={styles.columnHeader}>
              <div>
                <h2 className={styles.columnTitle}>{column.title}</h2>
                <p className={styles.columnSubtitle}>{column.subtitle}</p>
              </div>
              <span className={styles.count}>{cardsByStatus(column.key).length}</span>
            </header>

            <div className={styles.stack}>
              {cardsByStatus(column.key).map((card) => {
                const isOpen = openCardId === card.id;
                return (
                  <article key={card.id} className={`${styles.card} ${isOpen ? styles.cardOpen : ""}`}>
                    <button
                      type="button"
                      className={styles.cardButton}
                      onClick={() => setOpenCardId(isOpen ? "" : card.id)}
                      aria-expanded={isOpen}
                      aria-controls={`kanban-details-${card.id}`}
                    >
                      <div className={styles.cardTop}>
                        <div className={styles.cardMetaLeft}>
                          <span className={styles.cardId}>{card.id}</span>
                          <span className={`${styles.statusBadge} ${styles[`status_${card.status}`]}`}>{statusLabels[card.status]}</span>
                        </div>
                        <span className={`${styles.priority} ${styles[`priority_${card.priority.toLowerCase()}`]}`}>{card.priority}</span>
                      </div>
                      <div className={styles.cardHeading}>
                        <h3 className={styles.cardTitle}>{card.title}</h3>
                        <span className={`${styles.expandIcon} ${isOpen ? styles.expandIconOpen : ""}`} aria-hidden="true">
                          v
                        </span>
                      </div>
                      <p className={styles.cardSummary}>{card.summary}</p>
                    </button>

                    {isOpen ? (
                      <div id={`kanban-details-${card.id}`} className={styles.cardDetails}>
                        <div className={styles.detailBlock}>
                          <span className={styles.footerLabel}>Objetivo</span>
                          <p className={styles.footerText}>{card.outcome}</p>
                        </div>
                        <div className={styles.detailBlock}>
                          <span className={styles.footerLabel}>O que precisa ser implementado</span>
                          <ul className={styles.detailList}>
                            {card.implementation.map((item) => (
                              <li key={item} className={styles.detailItem}>
                                {item}
                              </li>
                            ))}
                          </ul>
                        </div>
                        <div className={styles.detailLinks}>
                          <button type="button" className={styles.docLink} onClick={handleOpenBacklogDoc}>
                            Abrir backlog oficial
                          </button>
                          <span className={styles.docHint}>{card.id}</span>
                        </div>
                        {docOpenHint ? <p className={styles.docStatus}>{docOpenHint}</p> : null}
                        <div className={styles.cardFooter}>
                          <span className={styles.footerLabel}>Proximo passo</span>
                          <p className={styles.footerText}>{card.nextStep}</p>
                        </div>
                      </div>
                    ) : null}
                  </article>
                );
              })}
            </div>
          </section>
        ))}
      </div>
    </section>
  );
}
