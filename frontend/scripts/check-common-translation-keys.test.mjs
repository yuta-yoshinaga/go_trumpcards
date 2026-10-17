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

function fixture(source, translations = { actionLog: { view: 'text' } }) {
  const dir = mkdtempSync(join(tmpdir(), 'common-tc-keys-'));
  dirs.push(dir);
  for (const root of ['pages', 'components']) mkdirSync(join(dir, root), { recursive: true });
  mkdirSync(join(dir, 'i18n', 'locales', 'ja'), { recursive: true });
  mkdirSync(join(dir, 'i18n', 'locales', 'en'), { recursive: true });
  writeFileSync(join(dir, 'pages', 'Page.tsx'), source);
  writeFileSync(join(dir, 'i18n', 'locales', 'ja', 'common.json'), JSON.stringify(translations));
  writeFileSync(join(dir, 'i18n', 'locales', 'en', 'common.json'), JSON.stringify(translations));
  return dir;
}

function check(dir) {
  const result = spawnSync('bun', [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-common-translation-keys', () => {
  it('accepts nested and flat literal common keys', () => {
    const nested = fixture("const label = tc('actionLog.view');", { actionLog: { view: 'text' } });
    expect(check(nested).code).toBe(0);
    const flat = fixture("const label = tc('actionLog.view');", { 'actionLog.view': 'text' });
    for (const lang of ['ja', 'en'])
      writeFileSync(join(flat, 'i18n', 'locales', lang, 'common.json'), JSON.stringify({ 'actionLog.view': 'text' }));
    expect(check(flat).code).toBe(0);
  });

  it('checks keys with options, defaults, and static template literals', () => {
    const dir = fixture(
      "tc('actionLog.view', { count: 1 }); tc('actionLog.view', 'fallback'); tc(`actionLog.view`); tc('missing.tutorial', { ns: 'tutorial' });",
    );
    expect(check(dir).code).toBe(0);
  });

  it('reports a missing key and locale', () => {
    const result = check(fixture("const label = tc('button.actionLog');"));
    expect(result.code).toBe(1);
    expect(result.out).toContain("tc('button.actionLog')");
    expect(result.out).toContain('ja/common.json');
    expect(result.out).toContain('en/common.json');
  });
});
