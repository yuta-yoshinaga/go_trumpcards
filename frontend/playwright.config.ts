import { defineConfig, devices } from '@playwright/test';

// In CI we pre-build the server binary so the cold `go build` happens outside
// the webServer startup window; locally we fall back to `go run`. Either way
// the server must run from the repo root so it can serve the `public` dir.
const serverCommand = process.env.E2E_SERVER_BIN
  ? `cd .. && PORT=8080 ${process.env.E2E_SERVER_BIN}`
  : 'cd .. && PORT=8080 go run ./cmd/server';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  // The server keys game state by sessionId (the frontend mints one per page
  // load via crypto.randomUUID), so specs cannot collide on shared state and
  // can run in parallel. Measured on a 20-spec sample: 1m24s at 1 worker vs
  // 38s at 4. Pinned to 4 rather than left undefined so a runner with more
  // cores does not oversubscribe one Chromium per core against a single
  // backend.
  workers: process.env.CI ? 4 : undefined,
  reporter: 'html',
  timeout: 90_000,
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          // CI runners give Chromium a tiny /dev/shm; once it fills during the
          // long single-worker run the browser process crashes mid-test
          // ("Target page, context or browser has been closed" — see #2369).
          // Routing shared memory to /tmp instead removes the flaky crash.
          args: process.env.CI ? ['--disable-dev-shm-usage'] : [],
        },
      },
    },
  ],
  webServer: {
    command: serverCommand,
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    // A cold `go run` compiles the whole backend (174 games) and can exceed a
    // tight window on slow runners; 120s absorbs that even without the prebuilt
    // binary, eliminating the chronic "Timed out waiting from webServer" flake.
    timeout: 120_000,
  },
});
