import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const SCRIPT = join(process.cwd(), 'scripts/check-game-page-definition.mjs');
const MIGRATED = 'export const AcesUpPage = createGamePage(def);';
const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function run({
  pages = {
    AcesUpPage: MIGRATED,
    HeartsPage: 'export const HeartsPage = createGamePage(def);',
    OldPage: 'export default function OldPage() {}',
  },
  baseline = '# fixture\nOldPage.tsx\n',
  routes = "export const routes = [{ page: 'AcesUp' }, { page: 'Hearts' }, { page: 'Old' }, { page: 'New' }];\n",
} = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'game-page-definition-'));
  dirs.push(dir);
  const pagesDir = join(dir, 'pages');
  mkdirSync(pagesDir);
  for (const [name, source] of Object.entries(pages)) writeFileSync(join(pagesDir, `${name}.tsx`), source);
  const baselinePath = join(dir, 'baseline.txt');
  writeFileSync(baselinePath, baseline);
  const routesPath = join(dir, 'routes.ts');
  writeFileSync(routesPath, routes);
  const result = spawnSync('bun', [SCRIPT], {
    encoding: 'utf8',
    env: {
      ...process.env,
      GAME_PAGE_DEFINITION_PAGES: pagesDir,
      GAME_PAGE_DEFINITION_BASELINE: baselinePath,
      GAME_PAGE_DEFINITION_ROUTES: routesPath,
    },
  });
  return { status: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-game-page-definition', () => {
  it('passes a migrated and baseline fixture', () => {
    const result = run();
    expect(result.status).toBe(0);
    expect(result.out).toContain(
      'game-page-definition: OK (2 migrated, 1 in baseline, 3 pages, 0 non-routed, 0 wrappers)',
    );
  });

  const wrapper = "import { OldPage } from './OldPage';\nexport function VariantPage() { return <OldPage />; }\n";
  it('passes a wrapper omitted from the baseline', () => {
    const result = run({
      pages: { AcesUpPage: MIGRATED, VariantPage: wrapper, OldPage: 'Old' },
      baseline: '# fixture\nOldPage.tsx\n',
      routes: "export const routes = [{ page: 'AcesUp' }, { page: 'Variant' }, { page: 'Old' }];",
    });
    expect(result.status).toBe(0);
    expect(result.out).toContain('1 wrappers');
  });

  it('treats a sibling page import with its own JSX as hand-wired', () => {
    const result = run({
      pages: {
        AcesUpPage: MIGRATED,
        VariantPage: `import { OldPage } from './OldPage';\nexport function VariantPage() { return <main><h1>Own page</h1></main>; }`,
        OldPage: 'Old',
      },
      baseline: '# fixture\nOldPage.tsx\n',
      routes: "export const routes = [{ page: 'AcesUp' }, { page: 'Variant' }, { page: 'Old' }];",
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('VariantPage.tsx: new page must be built with createGamePage');
  });

  it('rejects a wrapper in the baseline', () => {
    const result = run({
      pages: { AcesUpPage: MIGRATED, VariantPage: wrapper, OldPage: 'Old' },
      baseline: '# fixture\nOldPage.tsx\nVariantPage.tsx\n',
      routes: "export const routes = [{ page: 'AcesUp' }, { page: 'Variant' }, { page: 'Old' }];",
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('variant wrapper of OldPage — remove from baseline');
  });

  it('rejects a wrapper that imports GamePageShell', () => {
    const result = run({
      pages: {
        AcesUpPage: MIGRATED,
        VariantPage: `${wrapper}import { GamePageShell } from '../components/GamePageShell';`,
        OldPage: 'Old',
      },
      routes: "export const routes = [{ page: 'AcesUp' }, { page: 'Variant' }, { page: 'Old' }];",
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('wrapper of OldPage imports GamePageShell');
  });

  it('rejects a non-routed page in the baseline', () => {
    const result = run({
      pages: { AcesUpPage: MIGRATED, OldPage: 'Old', LegalPage: 'Legal' },
      baseline: '# fixture\nLegalPage.tsx\nOldPage.tsx\n',
      routes: "export const routes = [{ page: 'AcesUp' }, { page: 'Old' }];",
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('LegalPage.tsx: not a game page — remove from baseline');
  });

  it('rejects a new hand-wired page', () => {
    const result = run({
      pages: { AcesUpPage: MIGRATED, NewPage: 'export default function NewPage() {}', OldPage: 'Old' },
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('new page must be built with createGamePage (ADR-0044)');
  });

  it('does not count a createGamePage mention outside the page export as migrated', () => {
    const result = run({
      pages: {
        AcesUpPage: MIGRATED,
        NewPage: '// createGamePage(def) later\nexport function NewPage() {}',
        OldPage: 'Old',
      },
    });
    expect(result.status).toBe(1);
    expect(result.out).toContain('NewPage.tsx: new page must be built with createGamePage');
  });

  it('rejects a migrated baseline page', () => {
    const result = run({ baseline: '# fixture\nAcesUpPage.tsx\nOldPage.tsx\n' });
    expect(result.status).toBe(1);
    expect(result.out).toContain('migrated — remove it from the baseline');
  });

  it('rejects a stale baseline entry', () => {
    const result = run({ baseline: '# fixture\nMissingPage.tsx\nOldPage.tsx\n' });
    expect(result.status).toBe(1);
    expect(result.out).toContain('stale baseline entry');
  });

  it('rejects unsorted or duplicate baseline lines', () => {
    for (const baseline of ['# fixture\nOldPage.tsx\nAcesUpPage.tsx\n', '# fixture\nOldPage.tsx\nOldPage.tsx\n']) {
      const result = run({ baseline });
      expect(result.status).toBe(1);
      expect(result.out).toContain('baseline not sorted or contains duplicate lines');
    }
  });

  it('passes against the real repository', () => {
    const result = spawnSync('bun', [SCRIPT], { encoding: 'utf8' });
    expect(result.status).toBe(0);
    expect(result.stdout).toContain('game-page-definition: OK');
  });
});
