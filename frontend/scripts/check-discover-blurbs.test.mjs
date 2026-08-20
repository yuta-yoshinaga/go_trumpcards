import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-discover-blurbs.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

function routes(pages) {
  const entries = pages
    .map(
      (page) => `      {
        path: '/${page.toLowerCase()}',
        page: '${page}',
      },`,
    )
    .join('\n');
  return `export const GAME_ROUTES = [\n  {\n    routes: [\n${entries}\n    ],\n  },\n];\n`;
}

/** ja/en blurb maps that are complete and distinct — the healthy baseline. */
function locales() {
  return {
    ja: {
      blurb: { blackjack: '手札を21に近づける定番のゲーム。', poker: '役の強さを競うカードゲーム。' },
      stretch_blurb: { blackjack: '基本戦略を覚えると勝率が上がる。', poker: 'ブラフの読み合いが醍醐味。' },
    },
    en: {
      blurb: { blackjack: 'Get as close to 21 as you dare.', poker: 'Bet on the strength of your hand.' },
      stretch_blurb: { blackjack: 'Basic strategy sharpens your edge.', poker: 'Reading a bluff is the real game.' },
    },
  };
}

function fixture({ pages = ['BlackJack', 'Poker'], ja, en } = {}) {
  const base = locales();
  const dir = mkdtempSync(join(tmpdir(), 'discover-blurbs-'));
  dirs.push(dir);
  const src = join(dir, 'src');
  mkdirSync(join(src, 'constants'), { recursive: true });
  mkdirSync(join(src, 'i18n/locales/ja'), { recursive: true });
  mkdirSync(join(src, 'i18n/locales/en'), { recursive: true });
  writeFileSync(join(src, 'constants/gameRoutes.ts'), routes(pages));
  writeFileSync(join(src, 'i18n/locales/ja/discover.json'), JSON.stringify(ja ?? base.ja));
  writeFileSync(join(src, 'i18n/locales/en/discover.json'), JSON.stringify(en ?? base.en));
  return src;
}

function check(src) {
  const r = spawnSync(process.execPath, [GUARD, src], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-discover-blurbs', () => {
  // Positive control. The wording rules below are heuristics, so the case that
  // matters most is the one where correct prose stays quiet.
  it('accepts complete, distinct, translated blurbs', () => {
    const r = check(fixture());
    expect(r.code).toBe(0);
    expect(r.out).toContain('discover-blurbs: OK');
  });

  it('rejects a game with no blurb', () => {
    const { ja, en } = locales();
    delete ja.blurb.poker;
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('MISSING');
  });

  // The wording checks. Each of these renders perfectly on the page.
  it('rejects two games sharing one blurb', () => {
    const { ja, en } = locales();
    ja.blurb.poker = ja.blurb.blackjack;
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('DUPLICATE');
    expect(r.out).toContain('poker');
  });

  it('rejects an en entry left identical to ja', () => {
    const { ja, en } = locales();
    en.blurb.poker = ja.blurb.poker;
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('UNTRANSLATED');
  });

  it('rejects placeholder text', () => {
    const { ja, en } = locales();
    ja.blurb.poker = 'TODO あとで書く';
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('PLACEHOLDER');
  });

  it('rejects a stub too short to be a description', () => {
    const { ja, en } = locales();
    ja.blurb.poker = '準備中';
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('TOO SHORT');
  });

  it('checks stretch_blurb as well as blurb', () => {
    const { ja, en } = locales();
    ja.stretch_blurb.poker = ja.stretch_blurb.blackjack;
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('stretch_blurb');
  });

  it('does not confuse a duplicate across sections with a duplicate within one', () => {
    // The same sentence used once as a blurb and once as a stretch_blurb is
    // odd but not the failure being hunted; only collisions inside a single
    // section mean two games share a description.
    const { ja, en } = locales();
    ja.stretch_blurb.blackjack = ja.blurb.blackjack;
    const r = check(fixture({ ja, en }));
    expect(r.code).toBe(0);
  });
});
