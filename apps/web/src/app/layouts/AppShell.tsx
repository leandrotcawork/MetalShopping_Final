import { type PropsWithChildren, type ReactNode, useMemo, useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";

import { useSession } from "@metalshopping/feature-auth-session";

import logoIco from "../../assets/logo_ico.jpg";
import logoMetalNobre from "../../assets/logo_metal_nobre.svg";
import styles from "./AppShell.module.css";

type NavItem = {
  to: string;
  label: string;
  badge?: string;
  icon: ReactNode;
};

type NavSection = {
  label: string;
  items: NavItem[];
};

function HomeIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M4 11.5 12 5l8 6.5" />
      <path d="M6.5 10.8V19h11V10.8" />
      <path d="M10 19v-5h4v5" />
    </svg>
  );
}

function ShoppingIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M5 7h14l-1.1 8.2a1.5 1.5 0 0 1-1.5 1.3H8.6a1.5 1.5 0 0 1-1.5-1.3L6 7Z" />
      <path d="M9 7V5.8A3 3 0 0 1 12 3a3 3 0 0 1 3 2.8V7" />
      <circle cx="9" cy="20" r="1.2" fill="currentColor" stroke="none" />
      <circle cx="15" cy="20" r="1.2" fill="currentColor" stroke="none" />
    </svg>
  );
}

function ProductsIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M12 3 20 7.5 12 12 4 7.5 12 3Z" />
      <path d="M20 7.5V16.5L12 21 4 16.5V7.5" />
      <path d="M12 12v9" />
    </svg>
  );
}

function AnalyticsIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M5 19V9" />
      <path d="M12 19V5" />
      <path d="M19 19v-7" />
      <path d="M3.5 19.5h17" />
    </svg>
  );
}

function SettingsIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M12 8.5A3.5 3.5 0 1 0 12 15.5 3.5 3.5 0 1 0 12 8.5Z" />
      <path d="M19 12a7 7 0 0 0-.1-1.2l2-1.6-2-3.5-2.4 1a7.4 7.4 0 0 0-2-.9L14 3h-4l-.5 2.8a7.4 7.4 0 0 0-2 .9l-2.4-1-2 3.5 2 1.6A7 7 0 0 0 5 12c0 .4 0 .8.1 1.2l-2 1.6 2 3.5 2.4-1a7.4 7.4 0 0 0 2 .9L10 21h4l.5-2.8a7.4 7.4 0 0 0 2-.9l2.4 1 2-3.5-2-1.6c.1-.4.1-.8.1-1.2Z" />
    </svg>
  );
}

function LogoutIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path d="M14 7 19 12 14 17" />
      <path d="M19 12H9" />
      <path d="M11 5H6a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h5" />
    </svg>
  );
}

const sections: NavSection[] = [
  {
    label: "Main",
    items: [
      { to: "/home", label: "Inicio", icon: <HomeIcon /> },
      { to: "/shopping", label: "Shopping de Precos", icon: <ShoppingIcon /> },
      { to: "/products", label: "Produtos", icon: <ProductsIcon />, badge: "Live" },
    ],
  },
  {
    label: "Intelligence",
    items: [{ to: "/analytics", label: "Analytics", icon: <AnalyticsIcon /> }],
  },
  {
    label: "System",
    items: [{ to: "/settings", label: "Configuracoes", icon: <SettingsIcon /> }],
  },
];

function SidebarLink(item: NavItem) {
  return (
    <NavLink key={item.to} to={item.to} className={({ isActive }) => `${styles.link} ${isActive ? styles.linkActive : ""}`.trim()}>
      <span className={styles.linkMain}>
        <span className={styles.linkIcon} aria-hidden>
          {item.icon}
        </span>
        <span className={styles.linkLabel}>{item.label}</span>
      </span>
      {item.badge ? <small className={styles.linkBadge}>{item.badge}</small> : null}
    </NavLink>
  );
}

export function AppShell(_props: PropsWithChildren) {
  const [expanded, setExpanded] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const navigate = useNavigate();
  const { logout, session } = useSession();

  const displayName = useMemo(() => {
    if (!session?.display_name) {
      return "MetalShopping";
    }
    return session.display_name;
  }, [session?.display_name]);

  const roleLabel = useMemo(() => {
    const role = session?.roles?.[0];
    if (!role) {
      return "Operational Surface";
    }

    return role
      .split("_")
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join(" ");
  }, [session?.roles]);

  const avatarLabel = useMemo(() => {
    const source = displayName.trim();
    if (source === "") {
      return "MS";
    }

    const parts = source.split(/\s+/).filter(Boolean);
    const initials = parts
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? "")
      .join("");
    return initials || "MS";
  }, [displayName]);

  async function handleLogout() {
    setIsLoggingOut(true);
    try {
      await logout();
      navigate("/login?manual=1", { replace: true });
    } finally {
      setIsLoggingOut(false);
    }
  }

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
              {section.items.map(SidebarLink)}
            </div>
          ))}
        </nav>

        <footer className={styles.footer}>
          <div className={styles.userAvatar}>{avatarLabel}</div>
          <div className={styles.userMeta}>
            <p className={styles.userName}>{displayName}</p>
            <p className={styles.userRole}>{roleLabel}</p>
          </div>
          <button
            type="button"
            className={styles.logout}
            title="Sair"
            disabled={isLoggingOut}
            onClick={() => {
              void handleLogout();
            }}
          >
            <span className={styles.logoutIcon} aria-hidden>
              <LogoutIcon />
            </span>
            <span className={styles.logoutLabel}>{isLoggingOut ? "Saindo" : "Sair"}</span>
          </button>
        </footer>
      </aside>

      <main className={styles.content}>
        <Outlet />
      </main>
    </div>
  );
}
