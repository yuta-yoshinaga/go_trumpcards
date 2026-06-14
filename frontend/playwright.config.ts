import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
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
    command: 'cd .. && PORT=8080 go run ./cmd/server',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
