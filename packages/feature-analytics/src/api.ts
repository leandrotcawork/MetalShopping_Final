import type { ServerCoreSdk } from "@metalshopping/sdk-runtime";
import type { AnalyticsHomeV1 } from "@metalshopping/sdk-types";

export type AnalyticsHomeApi = Pick<ServerCoreSdk["analytics"], "getHome">;
export type AnalyticsHomeResult = AnalyticsHomeV1;
