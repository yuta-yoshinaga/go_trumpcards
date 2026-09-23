import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-japanese-locale-translated.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

/**
 * Build the internal locale tree consumed by the real guard process.
 *
 * Every fixture carries one correctly translated pair. Both walks assert a floor on how
 * many keys they inspected, so a fixture made only of English values makes the condition-2
 * floor fire before any violation is printed -- the guard would then look like it passed
 * the wrong way round. The anchor keeps both denominators non-zero without being a
 * violation itself: it is Japanese, so condition 1 skips it, and it carries no label, so
 * condition 2 clears it.
 */
const ANCHOR = { ja: { anchor: 'アンカー' }, en: { anchor: 'Anchor' } };

function fixture(ja, en, filename = 'game.json') {
  const dir = mkdtempSync(join(tmpdir(), 'japanese-locale-translated-'));
  dirs.push(dir);
  for (const [language, given] of Object.entries({ ja, en })) {
    const values = { ...ANCHOR[language], ...given };
    mkdirSync(join(dir, 'internal', 'i18n', 'locales', language), { recursive: true });
    writeFileSync(join(dir, 'internal', 'i18n', 'locales', language, filename), JSON.stringify(values, null, 2));
  }
  return dir;
}

function check(dir) {
  const result = spawnSync('bun', [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-japanese-locale-translated', () => {
  it('passes translated Japanese text', () => {
    const result = check(fixture({ title: 'タイトル', action: '捕獲' }, { title: 'Title', action: 'Capture' }));
    expect(result.code).toBe(0);
    expect(result.out).toContain('japanese-locale-translated: OK');
  });

  it('rejects a non-Japanese value identical to English', () => {
    const result = check(fixture({ foo: 'Capture cards' }, { foo: 'Capture cards' }));
    expect(result.code).toBe(1);
    expect(result.out).toContain('foo');
  });

  it('exempts command syntax under help keys', () => {
    const result = check(fixture({ helpFoo: 'm t <col> t <col>' }, { helpFoo: 'm t <col> t <col>' }));
    expect(result.code).toBe(0);
  });

  it('exempts command syntax under help keys from the embedded English-label check', () => {
    const result = check(
      fixture(
        { promptRoundEndHelp: 'nr / nextround: 次のラウンドへ' },
        { promptRoundEndHelp: 'nr / nextround: next round' },
      ),
    );
    expect(result.code).toBe(0);
  });

  it('does not broaden the help exemption to ordinary keys', () => {
    const result = check(fixture({ fooLabel: 'Capture cards' }, { fooLabel: 'Capture cards' }));
    expect(result.code).toBe(1);
    expect(result.out).toContain('fooLabel');
  });

  it('rejects an English label embedded in Japanese text', () => {
    const result = check(fixture({ action: '捕獲 played={{played}}' }, { action: 'Capture played={{played}}' }));
    expect(result.code).toBe(1);
    expect(result.out).toContain('action');
  });

  it('rejects a lowercase English label followed by a colon', () => {
    const result = check(fixture({ action: '捕獲 owner:{{owner}}' }, { action: 'Capture owner:{{owner}}' }));
    expect(result.code).toBe(1);
    expect(result.out).toContain('action');
  });

  it('passes Japanese text whose equals sign is only in a placeholder', () => {
    const result = check(fixture({ action: '捕獲 出した札={{played}}' }, { action: 'Capture cards={{played}}' }));
    expect(result.code).toBe(0);
  });

  it('exempts CLI option keys from the embedded English-label check', () => {
    const result = check(fixture({ optBuy: 'buy=バイ' }, { optBuy: 'buy=Buy' }));
    expect(result.code).toBe(0);
  });

  it('exempts CPU keys and established Latin terms from embedded English-label checks', () => {
    const result = check(
      fixture(
        {
          outputTitleLowball: '5-Card Draw Poker [2-7 Lowball]',
          cpuStats: 'CPU: 手札{{hand}}枚',
          weisLine: '当ラウンドの Weis: チーム0 {{t0}}点',
        },
        {
          outputTitleLowball: '5-Card Draw Poker [2-7 Lowball]',
          cpuStats: 'CPU: {{hand}} cards',
          weisLine: 'Round Weis: team 0 {{t0}} points',
        },
      ),
    );
    expect(result.code).toBe(0);
  });

  it('exempts Schafkopf contractShort.wenz from identical-to-English checks', () => {
    const result = check(fixture({ 'contractShort.wenz': 'Wenz' }, { 'contractShort.wenz': 'Wenz' }, 'schafkopf.json'));
    expect(result.code).toBe(0);
  });

  it('rejects the same contract label outside the Schafkopf contractShort exemption', () => {
    const otherFile = check(fixture({ 'contractShort.wenz': 'Wenz' }, { 'contractShort.wenz': 'Wenz' }, 'skat.json'));
    expect(otherFile.code).toBe(1);
    expect(otherFile.out).toContain('contractShort.wenz');

    const otherKey = check(fixture({ contractWenz: 'Wenz' }, { contractWenz: 'Wenz' }, 'schafkopf.json'));
    expect(otherKey.code).toBe(1);
    expect(otherKey.out).toContain('contractWenz');
  });
});
