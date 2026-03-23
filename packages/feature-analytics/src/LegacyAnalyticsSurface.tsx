import { AppSessionProvider } from "./app/providers/AppProviders";
import { AnalyticsPage } from "./pages/analytics/AnalyticsPage";

export function LegacyAnalyticsSurface() {
  return (
    <AppSessionProvider>
      <AnalyticsPage />
    </AppSessionProvider>
  );
}
