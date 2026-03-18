import type { ServerCoreSdk } from "@metalshopping/generated-sdk";
import type { AuthSessionStateV1 } from "@metalshopping/generated-types";

export type AuthSessionApi = ServerCoreSdk["authSession"];

export type AuthSessionStatus =
  | "bootstrapping"
  | "authenticated"
  | "unauthenticated"
  | "starting_login";

export type AuthSessionContextValue = {
  status: AuthSessionStatus;
  session: AuthSessionStateV1 | null;
  errorMessage: string;
  login: (returnTo?: string) => void;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
};
