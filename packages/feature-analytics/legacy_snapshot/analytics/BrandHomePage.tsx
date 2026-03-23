import { AnalyticsConcentrationSection } from "./components/AnalyticsConcentrationSection";
import { AnalyticsEfficiencySection } from "./components/AnalyticsEfficiencySection";
import { AnalyticsPerformanceSection, type AnalyticsPerformanceItem } from "./components/AnalyticsPerformanceSection";
import { KpiCard } from "./components/products/overview/KpiCard";
import styles from "./brand_home.module.css";

type BrandRow = {
  name: string;
  skuCount: number;
  revenue6m: string;
  margin6m: string;
  grossMargin: string;
  capital: string;
  riskExposure: string;
  readyPct: number;
  cautionBlockedPct: number;
  qualityScore: number;
  scaleScore: number;
  capitalWeight: number;
  readiness: "READY" | "CAUTION" | "BLOCKED";
  deltaRevenue: string;
  deltaMargin: string;
  deltaUrgency: string;
  suggestedRole: "Acelerar" | "Defender" | "Recuperar" | "Racionalizar";
  topCause: string;
  share: string;
  gmroi: number;
};

const kpis = [
  { label: "Marcas ativas", value: "24", change: "+3 vs base", tone: "positive" as const, helpItems: ["Quantidade de marcas acompanhadas no portfólio executivo."] },
  { label: "Receita 6M", value: "R$ 182,4M", change: "Top 5 = 78%", tone: "positive" as const, helpItems: ["Receita acumulada do recorte principal de 6 meses."] },
  { label: "Margem 6M", value: "23,8%", change: "-0,9 pp recente", tone: "negative" as const, helpItems: ["Margem consolidada do portfólio de marcas."] },
  { label: "Capital exposto", value: "R$ 61,4M", change: "71% no Top 5", tone: "negative" as const, helpItems: ["Capital concentrado nas marcas com maior peso de estoque e exposição."] },
  { label: "% READY", value: "42%", change: "10 marcas acionáveis", tone: "positive" as const, helpItems: ["Percentual de marcas com prontidão para ação."] },
  { label: "% CAUTION/BLOCKED", value: "29%", change: "7 marcas travadas", tone: "negative" as const, helpItems: ["Marcas com risco operacional ou baixa prontidão."] },
];

const abcRows = [
  { band: "A", label: "Núcleo dominante", share: "78%", detail: "5 marcas concentram a maior parte do faturamento." },
  { band: "B", label: "Corredor relevante", share: "15%", detail: "6 marcas sustentam diversificação útil com peso intermediário." },
  { band: "C", label: "Cauda longa", share: "7%", detail: "13 marcas com peso baixo e seletividade maior de capital." },
];

const brands: BrandRow[] = [
  { name: "Atlas", skuCount: 1248, revenue6m: "R$ 41,9M", margin6m: "18,4%", grossMargin: "R$ 7,7M", capital: "R$ 17,2M", riskExposure: "R$ 21,8M", readyPct: 18, cautionBlockedPct: 62, qualityScore: 36, scaleScore: 88, capitalWeight: 96, readiness: "BLOCKED", deltaRevenue: "-4,3%", deltaMargin: "-2,1 pp", deltaUrgency: "+18 pts", suggestedRole: "Recuperar", topCause: "Compressão de margem", share: "23%", gmroi: 1.21 },
  { name: "Vértice", skuCount: 904, revenue6m: "R$ 32,6M", margin6m: "29,7%", grossMargin: "R$ 9,7M", capital: "R$ 9,3M", riskExposure: "R$ 7,1M", readyPct: 67, cautionBlockedPct: 14, qualityScore: 86, scaleScore: 82, capitalWeight: 68, readiness: "READY", deltaRevenue: "+9,4%", deltaMargin: "+1,8 pp", deltaUrgency: "-11 pts", suggestedRole: "Acelerar", topCause: "Expandir execução", share: "18%", gmroi: 3.48 },
  { name: "Krona", skuCount: 1106, revenue6m: "R$ 25,5M", margin6m: "22,1%", grossMargin: "R$ 5,6M", capital: "R$ 11,6M", riskExposure: "R$ 13,4M", readyPct: 31, cautionBlockedPct: 39, qualityScore: 52, scaleScore: 74, capitalWeight: 80, readiness: "CAUTION", deltaRevenue: "-1,9%", deltaMargin: "-0,8 pp", deltaUrgency: "+7 pts", suggestedRole: "Defender", topCause: "Bloqueios operacionais", share: "14%", gmroi: 1.92 },
  { name: "Lumina", skuCount: 682, revenue6m: "R$ 21,3M", margin6m: "27,8%", grossMargin: "R$ 5,9M", capital: "R$ 5,9M", riskExposure: "R$ 4,8M", readyPct: 71, cautionBlockedPct: 11, qualityScore: 79, scaleScore: 71, capitalWeight: 52, readiness: "READY", deltaRevenue: "+6,2%", deltaMargin: "+2,8 pp", deltaUrgency: "-9 pts", suggestedRole: "Acelerar", topCause: "Expandir cobertura", share: "12%", gmroi: 3.06 },
  { name: "Prisma", skuCount: 754, revenue6m: "R$ 19,6M", margin6m: "24,3%", grossMargin: "R$ 4,8M", capital: "R$ 7,1M", riskExposure: "R$ 8,2M", readyPct: 44, cautionBlockedPct: 27, qualityScore: 61, scaleScore: 66, capitalWeight: 58, readiness: "CAUTION", deltaRevenue: "+1,3%", deltaMargin: "+0,2 pp", deltaUrgency: "-1 pt", suggestedRole: "Defender", topCause: "Readiness parcial", share: "11%", gmroi: 2.14 },
  { name: "Forza", skuCount: 518, revenue6m: "R$ 11,4M", margin6m: "15,2%", grossMargin: "R$ 1,7M", capital: "R$ 6,7M", riskExposure: "R$ 8,9M", readyPct: 9, cautionBlockedPct: 58, qualityScore: 28, scaleScore: 54, capitalWeight: 46, readiness: "BLOCKED", deltaRevenue: "-3,8%", deltaMargin: "-1,6 pp", deltaUrgency: "+10 pts", suggestedRole: "Racionalizar", topCause: "Excesso de capital", share: "6%", gmroi: 0.98 },
];

const momentumNotes = [
  { title: "Melhor avanço", body: "Lumina e Vértice combinam alta qualidade econômica com melhora recente de prontidão." },
  { title: "Pressão estrutural", body: "Atlas ainda sustenta o faturamento, mas já virou caso claro de recuperação." },
  { title: "Faixa de decisão", body: "Krona e Prisma estão na fronteira entre defender e acelerar, dependendo de execução." },
];

function bubbleTone(readiness: BrandRow["readiness"]): string {
  if (readiness === "READY") return styles.bubbleReady;
  if (readiness === "CAUTION") return styles.bubbleCaution;
  return styles.bubbleBlocked;
}

export function BrandHomePage() {
  const avgGmroi = brands.reduce((acc, brand) => acc + brand.gmroi, 0) / brands.length;
  const performanceItems: AnalyticsPerformanceItem[] = brands.slice(0, 6).map((brand) => ({
    id: brand.name,
    name: brand.name,
    icon: brand.name.slice(0, 2).toUpperCase(),
    metrics: [
      { label: "Receita", value: brand.revenue6m, delta: brand.deltaRevenue, deltaTone: brand.deltaRevenue.startsWith("+") ? "positive" as const : "negative" as const },
      { label: "Margem", value: brand.margin6m, delta: brand.deltaMargin, deltaTone: brand.deltaMargin.startsWith("+") ? "positive" as const : "negative" as const },
      { label: "Participacao", value: brand.share, subValue: brand.suggestedRole },
    ],
  }));
  const efficiencyItems = [...brands]
    .sort((a, b) => b.gmroi - a.gmroi)
    .map((brand) => ({
      id: brand.name,
      name: brand.name,
      supportingText: `${brand.grossMargin} margem bruta | ${brand.capital} capital`,
      progressPct: Math.min(100, (brand.gmroi / 4) * 100),
      valueLabel: `${brand.gmroi.toFixed(2).replace(".", ",")}x`,
    }));

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <h1 className={styles.title}>Marca</h1>
          <p className={styles.subtitle}>Analytics por portfólio de marcas</p>
        </div>
        <div className={styles.headerActions}>
          <span className={styles.headerChip}>Janela 6M</span>
          <span className={styles.headerChip}>Curta 3M</span>
          <button type="button" className={styles.secondaryBtn}>Comparar snapshots</button>
        </div>
      </header>

      <section className={styles.kpiGrid6}>
        {kpis.map((kpi) => (
          <KpiCard
            key={kpi.label}
            label={kpi.label}
            value={kpi.value}
            change={kpi.change}
            tone={kpi.tone}
            helpItems={kpi.helpItems}
            showChangeArrow={false}
          />
        ))}
      </section>

      <section className={styles.twoCol}>
        <AnalyticsPerformanceSection
          title="Performance por marca"
          hint="Receita, margem e participação das marcas mais relevantes do portfólio."
          spotlight={{ label: "Top 6" }}
          items={performanceItems}
          solid
        />

        <AnalyticsConcentrationSection
          title="Concentração de receita"
          hint="Monitorar dependência em poucas marcas."
          riskBadge="Top 5 = 78%"
          riskTone="negative"
          stats={[
            { label: "Top 3", value: "55%" },
            { label: "Top 5", value: "78%", valueTone: "negative" },
            { label: "Base", value: "Mock" },
          ]}
          bars={[
            { label: "Top 3 marcas", valueText: "55%", percent: 55, tone: "top3" },
            { label: "Top 5 marcas", valueText: "78%", percent: 78, tone: "top5" },
            { label: "Top 10 marcas", valueText: "93%", percent: 93, tone: "top10" },
          ]}
          quote="O portfólio concentra valor nas mesmas marcas que concentram risco e capital."
        />
      </section>

      <section className={styles.twoCol}>
        <article className={styles.panel}>
          <div className={styles.panelHeadRow}>
            <div className={styles.panelHead}>
              <h2 className={styles.panelTitle}>Mapa de portfólio de marcas</h2>
              <p className={styles.panelSub}>Bubble map por escala, qualidade econômica, capital exposto e prontidão.</p>
            </div>
            <span className={styles.panelActionText}>Bubble Map</span>
          </div>
          <div className={styles.mapShell}>
            <span className={`${styles.quadrant} ${styles.q1}`}>Defender</span>
            <span className={`${styles.quadrant} ${styles.q2}`}>Acelerar</span>
            <span className={`${styles.quadrant} ${styles.q3}`}>Recuperar</span>
            <span className={`${styles.quadrant} ${styles.q4}`}>Racionalizar</span>
            <span className={styles.axisY}>Qualidade econômica</span>
            <span className={styles.axisX}>Escala</span>
            {brands.map((brand) => (
              <div
                key={brand.name}
                className={`${styles.bubble} ${bubbleTone(brand.readiness)}`}
                style={{ left: `${brand.scaleScore}%`, top: `${100 - brand.qualityScore}%`, width: `${brand.capitalWeight}px`, height: `${brand.capitalWeight}px` }}
              >
                <strong>{brand.name}</strong>
                <span>{brand.readiness}</span>
              </div>
            ))}
          </div>
        </article>

        <AnalyticsEfficiencySection
          title="GMROI das marcas"
          hint="Eficiência de capital por marca, no espírito de leitura de portfolio economics."
          summaryLabel={`Média ${avgGmroi.toFixed(2).replace(".", ",")}x`}
          items={efficiencyItems}
        />
      </section>

      <section className={styles.twoCol}>
        <article className={styles.panel}>
          <div className={styles.panelHead}>
            <h2 className={styles.panelTitle}>ABC de marcas</h2>
            <p className={styles.panelSub}>Representação executiva do impacto no faturamento.</p>
          </div>
          <div className={styles.abcList}>
            {abcRows.map((row) => (
              <div key={row.band} className={styles.abcCard}>
                <div className={styles.abcBand}>{row.band}</div>
                <div>
                  <strong>{row.label}</strong>
                  <p>{row.detail}</p>
                </div>
                <strong className={styles.abcShare}>{row.share}</strong>
              </div>
            ))}
          </div>
        </article>

        <article className={styles.panel}>
          <div className={styles.panelHead}>
            <h2 className={styles.panelTitle}>Momentum 6M vs 3M</h2>
            <p className={styles.panelSub}>Leitura executiva resumida das mudanças recentes.</p>
          </div>
          <div className={styles.momentumList}>
            {momentumNotes.map((note) => (
              <article key={note.title} className={styles.momentumCard}>
                <strong>{note.title}</strong>
                <p>{note.body}</p>
              </article>
            ))}
          </div>
        </article>
      </section>

      <article className={styles.panel}>
        <div className={styles.panelHead}>
          <h2 className={styles.panelTitle}>Portfólio de marcas</h2>
          <p className={styles.panelSub}>Tabela mestre pronta para plugar a view model real depois.</p>
        </div>
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Marca</th>
                <th>SKUs</th>
                <th>Receita 6M</th>
                <th>Margem 6M</th>
                <th>Capital exposto</th>
                <th>Exposure ajustada ao risco</th>
                <th>% READY</th>
                <th>% CAUTION/BLOCKED</th>
                <th>Qualidade econômica</th>
                <th>Delta urgência</th>
                <th>Top causa</th>
                <th>Papel sugerido</th>
              </tr>
            </thead>
            <tbody>
              {brands.map((brand) => (
                <tr key={brand.name}>
                  <td><strong>{brand.name}</strong><span>{brand.readiness}</span></td>
                  <td>{brand.skuCount.toLocaleString("pt-BR")}</td>
                  <td>{brand.revenue6m}</td>
                  <td>{brand.margin6m}</td>
                  <td>{brand.capital}</td>
                  <td>{brand.riskExposure}</td>
                  <td>{brand.readyPct}%</td>
                  <td>{brand.cautionBlockedPct}%</td>
                  <td>{brand.qualityScore}</td>
                  <td>{brand.deltaUrgency}</td>
                  <td>{brand.topCause}</td>
                  <td><span className={styles.rolePill}>{brand.suggestedRole}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </article>
    </section>
  );
}
