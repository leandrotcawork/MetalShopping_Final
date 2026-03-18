import type { PropsWithChildren, ReactNode } from "react";
import { useState } from "react";

import styles from "./AppShell.module.css";

type NavItem = {
  key: string;
  label: string;
  badge?: string;
  icon: ReactNode;
  soon?: boolean;
};

type NavSection = {
  label: string;
  items: NavItem[];
};

type AppShellProps = PropsWithChildren<{
  activeItemKey: string;
}>;

function HomeIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.iconSvg}>
      <path d="M4 11.5 12 5l8 6.5V20a1 1 0 0 1-1 1h-4.5v-5.5h-5V21H5a1 1 0 0 1-1-1z" />
    </svg>
  );
}

function ProductsIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.iconSvg}>
      <path d="M5 7.5h14v11A1.5 1.5 0 0 1 17.5 20h-11A1.5 1.5 0 0 1 5 18.5z" />
      <path d="M8 7.5V6a4 4 0 0 1 8 0v1.5" />
    </svg>
  );
}

function ShoppingIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.iconSvg}>
      <path d="M4 6.5h2.4l1.6 8h9.1l1.6-6H8.8" />
      <circle cx="10" cy="18.5" r="1.5" />
      <circle cx="17" cy="18.5" r="1.5" />
    </svg>
  );
}

function AnalyticsIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.iconSvg}>
      <path d="M5 20V11.5M12 20V5M19 20v-8" />
    </svg>
  );
}

function SettingsIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.iconSvg}>
      <path d="m12 3 2 1.2 2.3-.3.9 2.1 2 1.2-.6 2.3.6 2.3-2 1.2-.9 2.1-2.3-.3L12 21l-2-1.2-2.3.3-.9-2.1-2-1.2.6-2.3-.6-2.3 2-1.2.9-2.1 2.3.3z" />
      <circle cx="12" cy="12" r="3.2" />
    </svg>
  );
}

const sections: NavSection[] = [
  {
    label: "Main",
    items: [
      { key: "home", label: "Inicio", icon: <HomeIcon />, soon: true },
      { key: "shopping", label: "Shopping", icon: <ShoppingIcon />, soon: true },
      { key: "products", label: "Produtos", icon: <ProductsIcon />, badge: "Live" },
    ],
  },
  {
    label: "Intelligence",
    items: [{ key: "analytics", label: "Analytics", icon: <AnalyticsIcon />, soon: true }],
  },
  {
    label: "System",
    items: [{ key: "settings", label: "Configuracoes", icon: <SettingsIcon />, soon: true }],
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

          <div className={styles.brandMark} aria-hidden="true">
            <span className={styles.brandMarkMetal}>M</span>
            <span className={styles.brandMarkShopping}>S</span>
          </div>

          <div className={styles.brandCopy}>
            <p className={styles.brandName}>
              <span className={styles.brandMetal}>Metal</span>
              <span className={styles.brandShopping}>Shopping</span>
            </p>
            <p className={styles.brandByline}>by Metal Nobre Acabamentos</p>
          </div>
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
                      <span className={styles.linkIcon}>{item.icon}</span>
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
