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
      title="Relatório de preço de mercado por run"
      subtitle="Selecione produtos por filtros e exporte um XLSX comparativo com preço interno versus concorrentes."
      aside={
        <div className={styles.heroAside}>
          <MetricChip label="Na grade">{props.totalVisible}</MetricChip>
          <MetricChip label="Selecionados">{props.totalSelected}</MetricChip>
          <MetricChip label="Total base">{props.totalProducts}</MetricChip>
          <MetricChip label="Runs">{props.totalRuns}</MetricChip>
          <div className={styles.heroActions}>
            <Button className={styles.actionButton} variant="secondary" onClick={props.onConfigureReport}>
              Configurar relatório
            </Button>
            <Button
              className={styles.actionButtonPrimary}
              variant="primary"
              disabled={props.exportDisabled}
              onClick={props.onExportReport}
            >
              Exportar relatório
            </Button>
          </div>
        </div>
      }
    />
  );
}
