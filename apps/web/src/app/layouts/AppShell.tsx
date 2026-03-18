import type { PropsWithChildren } from "react";
import { useState } from "react";

import logoIco from "../../assets/logo_ico.jpg";
import logoMetalNobre from "../../assets/logo_metal_nobre.svg";
import styles from "./AppShell.module.css";

type NavItem = {
  key: string;
  label: string;
  badge?: string;
  icon: string;
  soon?: boolean;
};

type NavSection = {
  label: string;
  items: NavItem[];
};

type AppShellProps = PropsWithChildren<{
  activeItemKey: string;
}>;

const sections: NavSection[] = [
  {
    label: "Main",
    items: [
      { key: "home", label: "Inicio", icon: "🏠", soon: true },
      { key: "shopping", label: "Shopping de Precos", icon: "🛒", soon: true },
      { key: "products", label: "Produtos", icon: "📦", badge: "Live" },
    ],
  },
  {
    label: "Intelligence",
    items: [{ key: "analytics", label: "Analytics", icon: "📊", soon: true }],
  },
  {
    label: "System",
    items: [{ key: "settings", label: "Configuracoes", icon: "⚙️", soon: true }],
  },
];

export function AppShell({ activeItemKey, children }: AppShellProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className={`${styles.shell} ${expanded ? styles.shellExpanded : ""}`.trim()}>
      <aside className={styles.sidebar}>
        <div className={styles.brandBlock}>
          <button
            type="button"
            className={styles.toggle}
            onClick={() => setExpanded((current) => !current)}
            aria-pressed={expanded}
            aria-label={expanded ? "Recolher menu" : "Expandir menu"}
            title={expanded ? "Recolher menu" : "Expandir menu"}
          >
            {expanded ? "<" : ">"}
          </button>
          <img src={logoIco} alt="MetalShopping" className={styles.brandIcon} />
          <img src={logoMetalNobre} alt="Metal Nobre Acabamentos" className={styles.brandLogo} />
          <p className={styles.brandName}>
            <span className={styles.brandMetal}>Metal</span>
            <span className={styles.brandShopping}>Shopping</span>
          </p>
          <p className={styles.brandByline}>by Metal Nobre Acabamentos</p>
        </div>

        <div className={styles.divider} />

        <nav className={styles.nav} aria-label="Primary">
          {sections.map((section) => (
            <div key={section.label} className={styles.section}>
              <p className={styles.sectionLabel}>{section.label}</p>
              {section.items.map((item) => {
                const active = item.key === activeItemKey;
                return (
                  <button
                    key={item.key}
                    type="button"
                    className={`${styles.link} ${active ? styles.linkActive : ""}`.trim()}
                    aria-current={active ? "page" : undefined}
                    disabled={item.soon && !active}
                    title={item.soon && !active ? "Coming soon" : item.label}
                  >
                    <span className={styles.linkMain}>
                      <span className={`${styles.linkIcon} ${styles.linkEmoji}`} aria-hidden>{item.icon}</span>
                      <span className={styles.linkLabel}>{item.label}</span>
                    </span>
                    {item.badge ? <small className={styles.linkBadge}>{item.badge}</small> : null}
                  </button>
                );
              })}
            </div>
          ))}
        </nav>

        <footer className={styles.footer}>
          <div className={styles.userAvatar}>MS</div>
          <div className={styles.userMeta}>
            <p className={styles.userName}>MetalShopping</p>
            <p className={styles.userRole}>Operational Surface</p>
          </div>
        </footer>
      </aside>

      <main className={styles.content}>{children}</main>
    </div>
  );
}
