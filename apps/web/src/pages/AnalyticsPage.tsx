import { AnalyticsHomePage } from "@metalshopping/feature-analytics";
import { Component, type ErrorInfo, type ReactNode } from "react";
import { AppFrame } from "@metalshopping/ui";

class AnalyticsErrorBoundary extends Component<{ children: ReactNode }, { error: Error | null }> {
  state: { error: Error | null } = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // Keep browser console signal for fast diagnosis in dev/prod.
    // eslint-disable-next-line no-console
    console.error("ANALYTICS_ROUTE_CRASH", error, info);
  }

  render() {
    if (!this.state.error) return this.props.children;
    const message = this.state.error?.message || String(this.state.error);
    return (
      <AppFrame
        eyebrow="Metal Analytics"
        title="Falha ao carregar Analytics"
        subtitle="A rota quebrou durante o render. Veja detalhes abaixo e console para stack trace."
      >
        <pre style={{ whiteSpace: "pre-wrap" }}>{message}</pre>
      </AppFrame>
    );
  }
}

export function AnalyticsPage() {
  return (
    <AnalyticsErrorBoundary>
      <AnalyticsHomePage />
    </AnalyticsErrorBoundary>
  );
}
