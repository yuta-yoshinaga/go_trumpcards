import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Same resolution note as check-design-doc-identifiers.test.mjs: vitest serves
// test modules under a non-file URL, so resolve the guard from cwd (the vitest
// root is `frontend/`) and assert it is there -- otherwise every case below
// silently becomes a spawn of a missing file.
const GUARD = join(process.cwd(), 'scripts', 'check-discover-axes.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

const AXES = `
export const PROFILE_MAX = 5;
export const AXES = {
  mood: {
    labelI18nKey: 'axis.mood.label',
    profileLength: 2,
    questions: [
      {
        questionI18nKey: 'axis.mood.q1',
        options: [
          { key: 'calm', i18nKey: 'option.mood.calm', profileIdx: 0 },
          { key: 'lively', i18nKey: 'option.mood.lively', profileIdx: 1 },
        ],
      },
    ],
  },
};
`;

/** A locale object covering every key the fixture axes declare. */
function fullLocale() {
  return {
    axis: { mood: { label: '雰囲気', q1: 'どんな気分？' } },
    option: { mood: { calm: '静か', lively: 'にぎやか' } },
  };
}

function routes(profiles) {
  const entries = profiles
    .map(
      ([page, body]) => `      {
        path: '/${page.toLowerCase()}',
        page: '${page}',
        profile: { ${body} },
      },`,
    )
    .join('\n');
  return `export const GAME_ROUTES = [\n  {\n    routes: [\n${entries}\n    ],\n  },\n];\n`;
}

function fixture({ axes = AXES, ja = fullLocale(), en = fullLocale(), profiles = [['BlackJack', 'mood: [3, 2]']] }) {
  const dir = mkdtempSync(join(tmpdir(), 'discover-axes-'));
  dirs.push(dir);
  const src = join(dir, 'src');
  mkdirSync(join(src, 'constants'), { recursive: true });
  mkdirSync(join(src, 'i18n/locales/ja'), { recursive: true });
  mkdirSync(join(src, 'i18n/locales/en'), { recursive: true });
  writeFileSync(join(src, 'constants/discoverAxes.ts'), axes);
  writeFileSync(join(src, 'constants/gameRoutes.ts'), routes(profiles));
  writeFileSync(join(src, 'i18n/locales/ja/discover.json'), JSON.stringify(ja));
  writeFileSync(join(src, 'i18n/locales/en/discover.json'), JSON.stringify(en));
  return src;
}

function check(src) {
  const r = spawnSync(process.execPath, [GUARD, src], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-discover-axes', () => {
  // Positive control first. Every case below asserts the guard *fires*; without
  // this one, a guard that always failed would look equally healthy.
  it('accepts definitions whose keys resolve and whose profiles fit', () => {
    const r = check(fixture({}));
    expect(r.code).toBe(0);
    expect(r.out).toContain('discover-axes: OK');
  });

  it('rejects an option key missing from ja', () => {
    const ja = fullLocale();
    delete ja.option.mood.lively;
    const r = check(fixture({ ja }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('option.mood.lively');
  });

  it('rejects an axis label missing from en only', () => {
    const en = fullLocale();
    delete en.axis.mood.label;
    const r = check(fixture({ en }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('axis.mood.label');
  });

  it('rejects an empty translation, which renders as blank rather than as a miss', () => {
    const ja = fullLocale();
    ja.option.mood.calm = '   ';
    const r = check(fixture({ ja }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('EMPTY KEY');
  });

  it('rejects a profile vector shorter than profileLength', () => {
    const r = check(fixture({ profiles: [['BlackJack', 'mood: [3]']] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('WRONG LENGTH');
  });

  it('rejects a profile value above PROFILE_MAX', () => {
    const r = check(fixture({ profiles: [['BlackJack', 'mood: [3, 9]']] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('OUT OF RANGE');
  });

  it('rejects a negative profile value', () => {
    const r = check(fixture({ profiles: [['BlackJack', 'mood: [3, -1]']] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('OUT OF RANGE');
  });

  it('rejects a game missing an axis entirely', () => {
    const r = check(fixture({ profiles: [['BlackJack', 'theme: [1, 1]']] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('MISSING AXIS');
  });

  it('rejects an axis that discoverAxes.ts does not define', () => {
    const r = check(fixture({ profiles: [['BlackJack', 'mood: [3, 2], vibes: [1]']] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('UNKNOWN AXIS');
  });

  it('reports every offending game, not just the first', () => {
    const r = check(
      fixture({
        profiles: [
          ['BlackJack', 'mood: [3]'],
          ['Poker', 'mood: [9, 9]'],
        ],
      }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('BlackJack');
    expect(r.out).toContain('Poker');
  });

  it('fails loudly when PROFILE_MAX cannot be found', () => {
    // A guard that cannot parse its own inputs must say so rather than pass.
    const r = check(fixture({ axes: AXES.replace('export const PROFILE_MAX = 5;', '') }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('PROFILE_MAX');
  });
});
