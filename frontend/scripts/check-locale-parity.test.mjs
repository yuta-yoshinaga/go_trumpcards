import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Same resolution note as the other guard tests: vitest serves test modules
// under a non-file URL, so resolve the guard from cwd (the vitest root is
// `frontend/`) and assert it is there -- otherwise every case below silently
// becomes a spawn of a missing file.
const GUARD = join(process.cwd(), 'scripts', 'check-locale-parity.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** Build a locales tree: { ja: {file: obj}, en: {file: obj} }. */
function fixture({ ja = {}, en = {} }) {
  const dir = mkdtempSync(join(tmpdir(), 'locale-parity-'));
  dirs.push(dir);
  for (const [lang, files] of [
    ['ja', ja],
    ['en', en],
  ]) {
    mkdirSync(join(dir, lang), { recursive: true });
    for (const [name, obj] of Object.entries(files)) {
      writeFileSync(join(dir, lang, name), JSON.stringify(obj, null, 2));
    }
  }
  return dir;
}

function check(dir) {
  const r = spawnSync(process.execPath, [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-locale-parity', () => {
  // Positive control first. Without it, a guard that rejected everything would
  // look just as healthy as one that works.
  it('passes when the two trees agree, including nested keys', () => {
    const { code, out } = check(
      fixture({
        ja: { 'a.json': { title: 'タイトル', header: { turn: '手番' } } },
        en: { 'a.json': { title: 'Title', header: { turn: 'Turn' } } },
      }),
    );
    expect(code).toBe(0);
    expect(out).toContain('locale-parity: OK');
  });

  it('rejects a key that only ja has, naming it', () => {
    const { code, out } = check(
      fixture({
        ja: { 'a.json': { title: 'タイトル', bidQuota: 'ノルマ' } },
        en: { 'a.json': { title: 'Title' } },
      }),
    );
    expect(code).toBe(1);
    expect(out).toContain('missing from en');
    expect(out).toContain('bidQuota');
  });

  it('rejects a key that only en has', () => {
    const { code, out } = check(
      fixture({ ja: { 'a.json': { title: 'タイトル' } }, en: { 'a.json': { title: 'Title', extra: 'x' } } }),
    );
    expect(code).toBe(1);
    expect(out).toContain('missing from ja');
    expect(out).toContain('extra');
  });

  it('rejects a file that exists in only one language', () => {
    const { code, out } = check(
      fixture({ ja: { 'a.json': { t: 'あ' }, 'b.json': { t: 'い' } }, en: { 'a.json': { t: 'a' } } }),
    );
    expect(code).toBe(1);
    expect(out).toContain('b.json');
    expect(out).toContain('not en');
  });

  // ja needs one plural form and en needs two; that is correct, not a gap.
  it('folds i18next plural suffixes onto their base key', () => {
    const { code, out } = check(
      fixture({
        ja: { 'a.json': { deckUnit: '{{count}}組' } },
        en: { 'a.json': { deckUnit_one: '{{count}} deck', deckUnit_other: '{{count}} decks' } },
      }),
    );
    expect(code).toBe(0);
    expect(out).toContain('locale-parity: OK');
  });
});
