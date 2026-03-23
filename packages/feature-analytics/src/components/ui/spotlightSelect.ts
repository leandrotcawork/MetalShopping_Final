export function createSpotlightSelectClassNames(overrides?: { wrap?: string }) {
  return {
    wrap: overrides?.wrap || "",
  };
}
