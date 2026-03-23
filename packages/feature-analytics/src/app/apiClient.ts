export class ApiClientError extends Error {
  status?: number;
  constructor(message: string, status?: number) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
  }
}

export type WorkspaceInsightsV2 = Record<string, unknown>;
export type WorkspaceInsightsRecommendationItemV2 = Record<string, unknown>;
