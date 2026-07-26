import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// https://vite.dev/config/
export default defineConfig({
  plugins: [tailwindcss(), react()],
  build: {
    outDir: '../public',
    emptyOutDir: false,
  },
  server: {
    fs: {
      // `src/constants/cuiManualTexts.ts` imports the CUI manuals from the
      // repo-root `docs/` tree via `?raw`. Vite 7.3.2+ denies serving files
      // outside the project root by default, which made every test that
      // reaches a game page fail to collect ("Denied ID .../docs/manual/...").
      // Allow the repo root explicitly; production builds inline these at
      // build time and are unaffected.
      allow: ['..'],
    },
  },
  test: {
    environment: 'happy-dom',
    setupFiles: './src/test/setup.ts',
    testTimeout: 10000,
    exclude: ['e2e/**', 'node_modules/**'],
    // Forked test workers inherit Node's default ~4 GB heap cap, which the
    // largest single file (NavBar.test.tsx, 489 tests) exceeds — the worker
    // dies with "Reached heap limit" and Vitest then SKIPS the ~200 tests
    // queued behind it while still reporting the run as green, which has
    // already let real failures reach CI. Machine RAM is not the constraint
    // (the cap is per process); this only raises the ceiling, it allocates
    // nothing up front. Keep CI's NODE_OPTIONS too: execArgv reaches only the
    // forked workers, while the coverage run also needs headroom in the PARENT
    // process that aggregates it.
    execArgv: ['--max-old-space-size=8192'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/main.tsx'],
    },
  },
});
