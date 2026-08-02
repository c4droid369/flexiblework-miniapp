import { defineConfig, devices } from '@playwright/test';

// E2E assumes the full stack is already serving:
//   - backend  http://localhost:8080
//   - frontend http://localhost:8081
// Bring it up with `docker compose up -d` from the repo root.

const PORT = Number(process.env.E2E_FRONTEND_PORT ?? 8081);
const baseURL = `http://localhost:${PORT}`;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false, // shared admin user, simpler serial flow
  retries: 0,
  workers: 1,
  reporter: 'list',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  use: {
    baseURL,
    headless: true,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
