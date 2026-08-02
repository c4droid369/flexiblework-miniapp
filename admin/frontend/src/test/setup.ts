// Vitest setup: shared polyfills and global mocks. Imported once per test file
// via vitest.config.ts setupFiles.

// Replace localStorage with an in-memory shim so multiple test files (and
// happy-dom's internal one) all see the same store. happy-dom's per-origin
// storage gets reset between tests; this shim persists for the file and
// gives us a deterministic surface for token assertions.
const mem = new Map<string, string>();
const shim = {
  getItem: (k: string) => mem.get(k) ?? null,
  setItem: (k: string, v: string) => mem.set(k, v),
  removeItem: (k: string) => mem.delete(k),
  clear: () => mem.clear(),
  key: (i: number) => Array.from(mem.keys())[i] ?? null,
  get length() {
    return mem.size;
  },
};
Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  writable: true,
  value: shim,
});

// Suppress noisy Element Plus unstyled warnings during unit tests — they
// fire when components render outside their app context (e.g., in unit
// tests where we only render a subcomponent).
const _origWarn = console.warn;
console.warn = (...args: unknown[]) => {
  const msg = String(args[0] ?? '');
  if (msg.includes('[Element Plus]') || msg.includes('Failed to resolve component')) return;
  _origWarn(...(args as []));
};
