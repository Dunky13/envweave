/// <reference types="astro/client" />

// The PostHog install stub attaches `posthog` to `window`. Typed to the shape
// the consent bootstrap actually calls, so the processed script type-checks
// without `any`.
interface Window {
  posthog: {
    init(token: string, config: Record<string, unknown>): void;
    capture(event: string, properties?: Record<string, unknown>): void;
  };
}
