import { defineConfig } from "vitest/config";

const isWindows = process.platform === "win32";

export default defineConfig({
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
