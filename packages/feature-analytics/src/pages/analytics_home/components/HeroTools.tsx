// @ts-nocheck
import { useEffect, useRef, useState } from "react";

import type { AnalyticsHomeViewModel } from "../analyticsHomeViewModel";
import { AnimatedNumber } from "./AnimatedNumber";
import styles from "../analytics_home.module.css";

type HeroToolsProps = {
  onOpenSpotlight: (key: string) => void;
  onFilterChange: (filter: FilterKey) => void;
  miniStats: AnalyticsHomeViewModel["miniStats"];
};

type FilterKey = "all" | "critical" | "pricing" | "stock" | "data";

export function HeroTools({ onOpenSpotlight, onFilterChange, miniStats }: HeroToolsProps) {
  const [filter, setFilter] = useState<FilterKey>("all");
  const [query, setQuery] = useState("");
  const searchRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        searchRef.current?.focus();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  function setActiveFilter(next: FilterKey) {
    setFilter(next);
    onFilterChange(next);
  }

  const visibleMiniStats = miniStats.filter(
    (item) => !["mini-potential", "mini-margin", "mini-abc", "mini-capital"].includes(String(item.key || ""))
  );

  return (
    <aside className={styles.heroTools}>
      <div className={styles.toolTop}>
        <div>
          <div className={styles.toolTitle}>Barra de Comando</div>
          <div className={styles.cardSub}>Buscar SKU, marca, classificacoes ou regra • atalhos no teclado</div>
        </div>
        <div className={styles.chipRow}>
          <div className={styles.chipLine}>
            <button type="button" className={`${styles.chip} ${filter === "all" ? styles.active : ""}`} onClick={() => setActiveFilter("all")}>Tudo</button>
            <button type="button" className={`${styles.chip} ${filter === "critical" ? styles.active : ""}`} onClick={() => setActiveFilter("critical")}>Critico</button>
            <button type="button" className={`${styles.chip} ${filter === "pricing" ? styles.active : ""}`} onClick={() => setActiveFilter("pricing")}>Preco</button>
          </div>
          <div className={styles.chipLine}>
            <button type="button" className={`${styles.chip} ${filter === "stock" ? styles.active : ""}`} onClick={() => setActiveFilter("stock")}>Estoque</button>
            <button type="button" className={`${styles.chip} ${filter === "data" ? styles.active : ""}`} onClick={() => setActiveFilter("data")}>Dados</button>
          </div>
        </div>
      </div>

      <div className={styles.search}>
        <span>{"\u{1F50D}"}</span>
        <input
          ref={searchRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Ex.: 18578 - PORCELANATOS - PORTOBELLO - preco acima do mercado..."
        />
        <span className={styles.hint}>Ctrl K</span>
      </div>

      <div className={styles.toolGrid}>
        {visibleMiniStats.map((item) => (
          <button key={item.key} type="button" className={styles.miniStat} onClick={() => onOpenSpotlight(item.key)}>
            <span className={styles.l}>
              <small className={styles.miniK}>{item.label}</small>
              <AnimatedNumber as="strong" className={styles.v} value={item.value} />
              <small className={styles.s}>{item.sub}</small>
            </span>
            <span className={styles.badge}>{item.badge}</span>
          </button>
        ))}
      </div>
    </aside>
  );
}

