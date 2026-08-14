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
    //
    // 2026-08-14: 317 ゲームで 8 GB を超え、shard 3 が OOM で落ちるようになった
    // (`node::OOMErrorHandler`)。ゲームが 1 本増えるたびに NavBar/DesktopSidebar の
    // カタログ検査が全ルートを描くので、**上限は件数に比例して効かなくなる** ──
    // runner は 16 GB あるので 12 GB へ上げる。次に当たったらシャード数を増やす
    // ほうが筋が良い (上限の引き上げは 3 度目)。
    execArgv: ['--max-old-space-size=12288'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/main.tsx'],
    },
  },
});
