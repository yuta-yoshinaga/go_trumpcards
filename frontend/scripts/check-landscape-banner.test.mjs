import { spawnSync } from 'node:child_process';
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const ORIGINAL_GUARD = join(process.cwd(), 'scripts', 'check-landscape-banner.mjs');
if (!existsSync(ORIGINAL_GUARD)) throw new Error(`guard not found at ${ORIGINAL_GUARD}`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

function fixture({ variants = { standard: 83, cruel: 1, extra: 0 } }) {
  const dir = mkdtempSync(join(tmpdir(), 'landscape-banner-'));
  dirs.push(dir);

  // To avoid modifying the guard script to accept arguments, we copy the script
  // and its dependencies into the fixture directory so import.meta.url resolves
  // to our fixture's src/i18n/locales.
  const scriptsDir = join(dir, 'scripts');
  mkdirSync(join(scriptsDir, 'lib'), { recursive: true });
  copyFileSync(ORIGINAL_GUARD, join(scriptsDir, 'check-landscape-banner.mjs'));
  copyFileSync(join(process.cwd(), 'scripts', 'lib', 'floor.mjs'), join(scriptsDir, 'lib', 'floor.mjs'));

  const src = join(dir, 'src');
  const locales = join(src, 'i18n', 'locales');
  mkdirSync(join(locales, 'ja'), { recursive: true });
  mkdirSync(join(locales, 'en'), { recursive: true });

  let gameId = 1;
  const addGames = (count, bannerText) => {
    for (let i = 0; i < count; i++) {
      const name = `game${gameId++}`;
      const content = JSON.stringify({ landscapeBanner: bannerText });
      writeFileSync(join(locales, 'ja', `${name}.json`), content);
      writeFileSync(join(locales, 'en', `${name}.json`), content);
    }
  };

  addGames(variants.standard, 'Standard Banner');
  addGames(variants.cruel, 'Cruel Banner');
  addGames(variants.extra, 'Extra Banner');

  return dir;
}

function check(dir) {
  const GUARD = join(dir, 'scripts', 'check-landscape-banner.mjs');
  const r = spawnSync(process.execPath, [GUARD], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-landscape-banner', () => {
  it('passes when exactly 2 variations exist (standard and cruel)', () => {
    const { code, out } = check(fixture({ variants: { standard: 83, cruel: 1, extra: 0 } }));
    expect(code).toBe(0);
    expect(out).toContain('checked landscapeBanner across');
  });

  it('fails when a 3rd variation exists, outputting the violating game name', () => {
    const { code, out } = check(fixture({ variants: { standard: 83, cruel: 1, extra: 1 } }));
    expect(code).toBe(1);
    expect(out).toContain('Found 3 variations');
    expect(out).toContain('Extra Banner');
  });

  it('fails when the total games count falls below the floor of 60', () => {
    const { code, out } = check(fixture({ variants: { standard: 50, cruel: 1, extra: 0 } }));
    expect(code).toBe(1);
    expect(out).toContain('only 51 games found');
  });
});
