import { NavLink, useLocation } from "react-router-dom";

import styles from "../product_workspace.module.css";

const TABS = [
  { to: "overview", label: "Visao Geral" },
  { to: "history", label: "Historico" },
  { to: "simulator", label: "Simulador" },
  { to: "insights", label: "Insights" },
];

type WorkspaceTabsProps = {
  className?: string;
};

export function WorkspaceTabs({ className }: WorkspaceTabsProps) {
  const location = useLocation();
  const navState =
    typeof location.state === "object" && location.state
      ? location.state
      : undefined;

  return (
    <nav className={`${styles.tabs}${className ? ` ${className}` : ""}`}>
      {TABS.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          replace
          state={navState}
          className={({ isActive }) => `${styles.tab}${isActive ? ` ${styles.tabActive}` : ""}`}
        >
          {tab.label}
        </NavLink>
      ))}
    </nav>
  );
}
