import { type MutableRefObject, type PropsWithChildren, type ReactNode, useEffect, useRef } from "react";
import { createPortal } from "react-dom";

import styles from "../analytics_home.module.css";

type AnalyticsSpotlightDrawerProps = PropsWithChildren<{
  open: boolean;
  title: string;
  meta?: string;
  headerChips?: ReactNode;
  bodyRef?: MutableRefObject<HTMLDivElement | null>;
  cardClassName?: string;
  bodyClassName?: string;
  onClose: () => void;
}>;

export function AnalyticsSpotlightDrawer({
  open,
  title,
  meta,
  headerChips,
  bodyRef,
  cardClassName,
  bodyClassName,
  onClose,
  children,
}: AnalyticsSpotlightDrawerProps) {
  const drawerRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return undefined;

    const body = document.body;
    const prevOverflow = body.style.overflow;
    const prevPaddingRight = body.style.paddingRight;
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;

    body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      body.style.paddingRight = `${scrollbarWidth}px`;
    }

    const previous = document.activeElement as HTMLElement | null;
    const rafId = window.requestAnimationFrame(() => {
      drawerRef.current?.focus();
    });
    return () => {
      window.cancelAnimationFrame(rafId);
      previous?.focus?.();
      body.style.overflow = prevOverflow;
      body.style.paddingRight = prevPaddingRight;
    };
  }, [open]);

  useEffect(() => {
    if (!open) return undefined;
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  if (!open || typeof document === "undefined") {
    return null;
  }

  return createPortal(
    <>
      <button
        type="button"
        aria-label="Fechar Spotlight"
        className={`${styles.backdrop} ${styles.open}`}
        onClick={onClose}
        tabIndex={0}
      />
      <aside
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Spotlight"
        aria-hidden={false}
        tabIndex={-1}
        className={`${styles.drawer} ${styles.open}`}
        onClick={(event) => {
          if (event.target === event.currentTarget) onClose();
        }}
      >
        <div className={`${styles.drawerCard}${cardClassName ? ` ${cardClassName}` : ""}`} onClick={(event) => event.stopPropagation()}>
          <header className={styles.drawerHead}>
            <div className={styles.drawerHeadMain}>
              <h3>{title}</h3>
              {meta ? <p>{meta}</p> : null}
            </div>
            {headerChips ? <div className={styles.drawerHeadChips}>{headerChips}</div> : null}
            <button type="button" className={styles.close} onClick={onClose} aria-label="Fechar Spotlight">
              x
            </button>
          </header>
          <div
            ref={(node) => {
              if (bodyRef) bodyRef.current = node;
            }}
            className={`${styles.drawerBody}${bodyClassName ? ` ${bodyClassName}` : ""}`}
          >
            {children}
          </div>
        </div>
      </aside>
    </>,
    document.body
  );
}
