import { AppFrame, Button, MetricChip } from "@metalshopping/ui";

import styles from "../ProductsPortfolioPage.module.css";

export function ProductsHero(props: {
  totalVisible: number;
  totalSelected: number;
  totalProducts: number;
  totalRuns: number;
  exportDisabled: boolean;
  onConfigureReport: () => void;
  onExportReport: () => void;
}) {
  return (
    <AppFrame
      fullWidth
      eyebrow="Products · Market Report"
      title="Relatorio de preco de mercado por run"
      subtitle="Escolha uma run e exporte um XLSX comparativo com os produtos observados nela e seus concorrentes."
      aside={
        <div className={styles.heroAside}>
          <MetricChip label="Na grade">{props.totalVisible}</MetricChip>
          <MetricChip label="Selecionados">{props.totalSelected}</MetricChip>
          <MetricChip label="Total base">{props.totalProducts}</MetricChip>
          <MetricChip label="Runs">{props.totalRuns}</MetricChip>
          <div className={styles.heroActions}>
            <Button className={styles.actionButton} variant="secondary" onClick={props.onConfigureReport}>
              Configurar relatorio
            </Button>
            <Button
              className={styles.actionButtonPrimary}
              variant="primary"
              disabled={props.exportDisabled}
              onClick={props.onExportReport}
            >
              Exportar relatorio
            </Button>
          </div>
        </div>
      }
    />
  );
}
