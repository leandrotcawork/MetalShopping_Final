import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import styles from "../product_workspace.module.css";

export type TooltipHelpCopy = {
  title: string;
  items: string[];
};

type InfoTooltipLabelProps = {
  label: string;
  help?: TooltipHelpCopy | null;
  className?: string;
};

export function InfoTooltipLabel({ label, help, className }: InfoTooltipLabelProps) {
  const tooltipId = useId();
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const tooltipRef = useRef<HTMLSpanElement | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [position, setPosition] = useState<{ left: number; top: number }>({ left: 0, top: 0 });

  const updatePosition = () => {
    if (!buttonRef.current || !tooltipRef.current) {
      return;
    }

    const anchorRect = buttonRef.current.getBoundingClientRect();
    const tooltipRect = tooltipRef.current.getBoundingClientRect();
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const spacing = 10;
    const margin = 12;

    let left = anchorRect.left + anchorRect.width / 2 - tooltipRect.width / 2;
    left = Math.max(margin, Math.min(left, viewportWidth - tooltipRect.width - margin));

    const topAbove = anchorRect.top - tooltipRect.height - spacing;
    const topBelow = anchorRect.bottom + spacing;
    const top =
      topAbove >= margin || topBelow + tooltipRect.height > viewportHeight - margin ? topAbove : topBelow;

    setPosition({ left, top: Math.max(margin, top) });
  };

  useLayoutEffect(() => {
    if (!isOpen) {
      return;
    }
    updatePosition();
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const onViewportChange = () => updatePosition();
    window.addEventListener("scroll", onViewportChange, true);
    window.addEventListener("resize", onViewportChange);
    return () => {
      window.removeEventListener("scroll", onViewportChange, true);
      window.removeEventListener("resize", onViewportChange);
    };
  }, [isOpen]);

  return (
    <span className={`${styles.labelWithInfo}${className ? ` ${className}` : ""}`}>
      <span>{label}</span>
      {help ? (
        <span className={styles.infoWrap}>
          <button
            ref={buttonRef}
            type="button"
            className={styles.infoButton}
            aria-label={`Info: ${help.title}`}
            aria-describedby={isOpen ? tooltipId : undefined}
            onMouseEnter={() => setIsOpen(true)}
            onMouseLeave={() => setIsOpen(false)}
            onFocus={() => setIsOpen(true)}
            onBlur={() => setIsOpen(false)}
          >
            <svg className={styles.infoIcon} viewBox="0 0 20 20" aria-hidden>
              <circle cx="10" cy="10" r="7.4" fill="none" stroke="currentColor" strokeWidth="2.1" />
              <circle cx="10" cy="6.35" r="1.5" fill="currentColor" />
              <line x1="10" y1="9.1" x2="10" y2="14.4" stroke="currentColor" strokeWidth="2.1" strokeLinecap="round" />
            </svg>
          </button>
          {isOpen && typeof document !== "undefined"
            ? createPortal(
                <span
                  id={tooltipId}
                  ref={tooltipRef}
                  className={styles.infoTooltipFloating}
                  style={{ left: `${position.left}px`, top: `${position.top}px` }}
                >
                  <span className={styles.infoTitle}>{help.title}</span>
                  <ul className={styles.infoList}>
                    {help.items.map((item) => (
                      <li key={item}>{item}</li>
                    ))}
                  </ul>
                </span>,
                document.body,
              )
            : null}
        </span>
      ) : null}
    </span>
  );
}
