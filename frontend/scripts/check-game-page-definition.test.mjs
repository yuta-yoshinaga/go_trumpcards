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
} = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'game-page-definition-'));
  dirs.push(dir);
  const pagesDir = join(dir, 'pages');
  mkdirSync(pagesDir);
  for (const [name, source] of Object.entries(pages)) writeFileSync(join(pagesDir, `${name}.tsx`), source);
  const baselinePath = join(dir, 'baseline.txt');
  writeFileSync(baselinePath, baseline);
  const result = spawnSync('bun', [SCRIPT], {
    encoding: 'utf8',
    env: { ...process.env, GAME_PAGE_DEFINITION_PAGES: pagesDir, GAME_PAGE_DEFINITION_BASELINE: baselinePath },
  });
  return { status: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-game-page-definition', () => {
  it('passes a migrated and baseline fixture', () => {
    const result = run();
    expect(result.status).toBe(0);
    expect(result.out).toContain('game-page-definition: OK (2 migrated, 1 in baseline, 3 pages)');
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
