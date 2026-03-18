import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@metalshopping/ui": path.resolve(__dirname, "../../packages/ui/src"),
      "@metalshopping/feature-products": path.resolve(
        __dirname,
        "../../packages/feature-products/src",
      ),
      "@metalshopping/generated-types": path.resolve(
        __dirname,
        "../../packages/generated/types_ts/contracts.generated.ts",
      ),
    },
  },
  server: {
    host: "127.0.0.1",
    port: 4173,
  },
  test: {
    include: [
      "src/**/*.test.ts",
      "src/**/*.test.tsx",
      "../../packages/feature-products/src/**/*.test.ts",
      "../../packages/feature-products/src/**/*.test.tsx",
    ],
  },
});
