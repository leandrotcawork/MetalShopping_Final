import { AppSessionProvider } from "./app/providers/AppProviders";
import { AnalyticsProductsWorkspacePage } from "./pages/analytics/AnalyticsProductsWorkspacePage";

export function LegacyAnalyticsProductsWorkspaceSurface() {
  return (
    <AppSessionProvider>
      <AnalyticsProductsWorkspacePage />
    </AppSessionProvider>
  );
}
