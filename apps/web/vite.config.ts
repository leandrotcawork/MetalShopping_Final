import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

const isWindows = process.platform === "win32";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@metalshopping/ui": path.resolve(__dirname, "../../packages/ui/src"),
      "@metalshopping/feature-auth-session": path.resolve(
        __dirname,
        "../../packages/feature-auth-session/src",
      ),
      "@metalshopping/feature-analytics": path.resolve(
        __dirname,
        "../../packages/feature-analytics/src",
      ),
      "@metalshopping/feature-products": path.resolve(
        __dirname,
        "../../packages/feature-products/src",
      ),
      "@metalshopping/sdk-runtime": path.resolve(
        __dirname,
        "../../packages/platform-sdk/src",
      ),
      "@metalshopping/sdk-types": path.resolve(
        __dirname,
        "../../packages/generated-types/src",
      ),
    },
  },
  server: {
    host: "127.0.0.1",
    port: 5173,
  },
  test: {
    include: [
      "src/**/*.test.ts",
      "src/**/*.test.tsx",
      "../../packages/feature-auth-session/src/**/*.test.ts",
      "../../packages/feature-auth-session/src/**/*.test.tsx",
      "../../packages/feature-analytics/src/**/*.test.ts",
      "../../packages/feature-analytics/src/**/*.test.tsx",
      "../../packages/feature-products/src/**/*.test.ts",
      "../../packages/feature-products/src/**/*.test.tsx",
    ],
    pool: isWindows ? "vmThreads" : "forks",
    poolOptions: {
      vmThreads: {
        useAtomics: true,
      },
    },
  },
});
