import { useNavigate } from "react-router-dom";

import { WorkspaceTabs } from "./WorkspaceTabs";
import styles from "../product_workspace.module.css";

type WorkspaceHeaderProps = {
  updatedAt: string;
  fromPath: string | null;
  fromScrollY: number | null;
};

export function WorkspaceHeader({ updatedAt, fromPath, fromScrollY }: WorkspaceHeaderProps) {
  const navigate = useNavigate();

  function handleBack() {
    if (fromPath) {
      navigate(fromPath, {
        state: fromScrollY != null ? { restoreScrollY: fromScrollY } : undefined,
      });
      return;
    }
    if (window.history.length > 1) {
      navigate(-1);
      return;
    }
    navigate("/analytics/products");
  }

  return (
    <header className={styles.header}>
      <div className={styles.headerContent}>
        <button type="button" className={styles.backBtn} onClick={handleBack} aria-label="Voltar para lista">
          <span aria-hidden>←</span>
        </button>
        <div className={styles.logo}>
          METAL <span>ANALYTICS</span>
        </div>
        <WorkspaceTabs />
        <div className={styles.updateTime}>Atualizado em: {updatedAt}</div>
      </div>
    </header>
  );
}
