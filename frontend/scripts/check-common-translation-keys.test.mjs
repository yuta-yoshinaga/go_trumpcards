import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-common-translation-keys.mjs');
const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(key) {
  const dir = mkdtempSync(join(tmpdir(), 'common-tc-keys-'));
  dirs.push(dir);
  for (const root of ['pages', 'components']) mkdirSync(join(dir, root), { recursive: true });
  for (let index = 0; index < 100; index += 1) {
    writeFileSync(join(dir, 'components', `Scanned${index}.tsx`), 'export {};');
  }
  mkdirSync(join(dir, 'i18n', 'locales', 'ja'), { recursive: true });
  mkdirSync(join(dir, 'i18n', 'locales', 'en'), { recursive: true });
  writeFileSync(join(dir, 'pages', 'Page.tsx'), `const label = tc('${key}');`);
  writeFileSync(
    join(dir, 'i18n', 'locales', 'ja', 'common.json'),
    JSON.stringify({ actionLog: { view: '棋譜を見る' } }),
  );
  writeFileSync(
    join(dir, 'i18n', 'locales', 'en', 'common.json'),
    JSON.stringify({ actionLog: { view: 'View action log' } }),
  );
  return dir;
}

function check(dir) {
  const result = spawnSync('bun', [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-common-translation-keys', () => {
  it('accepts nested and flat literal common keys', () => {
    const nested = fixture('actionLog.view');
    expect(check(nested).code).toBe(0);
    const flat = fixture('actionLog.view');
    for (const lang of ['ja', 'en'])
      writeFileSync(join(flat, 'i18n', 'locales', lang, 'common.json'), JSON.stringify({ 'actionLog.view': 'text' }));
    expect(check(flat).code).toBe(0);
  });

  it('reports a missing key and locale', () => {
    const result = check(fixture('button.actionLog'));
    expect(result.code).toBe(1);
    expect(result.out).toContain("tc('button.actionLog')");
    expect(result.out).toContain('ja/common.json');
    expect(result.out).toContain('en/common.json');
  });
});
