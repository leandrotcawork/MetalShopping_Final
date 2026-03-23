import { Navigate, Route, Routes } from "react-router-dom";

import { ProductWorkspaceLayout } from "./workspace/ProductWorkspaceLayout";
import { HistoryTab } from "./workspace/tabs/HistoryTab";
import { InsightsTab } from "./workspace/tabs/InsightsTab";
import { OverviewTab } from "./workspace/tabs/OverviewTab";
import { SimulatorTab } from "./workspace/tabs/SimulatorTab";

export function AnalyticsProductsWorkspacePage() {
  return (
    <Routes>
      <Route element={<ProductWorkspaceLayout />}>
        <Route index element={<Navigate replace to="overview" />} />
        <Route path="overview" element={<OverviewTab />} />
        <Route path="insights" element={<InsightsTab />} />
        <Route path="history" element={<HistoryTab />} />
        <Route path="simulator" element={<SimulatorTab />} />
        <Route path="*" element={<Navigate replace to="overview" />} />
      </Route>
    </Routes>
  );
}
