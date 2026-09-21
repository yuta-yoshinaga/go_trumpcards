import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-locale-text-agreement.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(cuiText, webText, { cuiLog, webLog, cuiEntries = {}, webEntries = {}, game = 'example' } = {}) {
  const root = mkdtempSync(join(tmpdir(), 'locale-text-agreement-'));
  dirs.push(root);
  mkdirSync(join(root, 'internal', 'i18n', 'locales', 'en'), { recursive: true });
  mkdirSync(join(root, 'frontend', 'src', 'i18n', 'locales', 'en'), { recursive: true });
  writeFileSync(
    join(root, 'internal', 'i18n', 'locales', 'en', `${game}.json`),
    JSON.stringify({ errExample: cuiText, ...(cuiLog ? { 'log.example': cuiLog } : {}), ...cuiEntries }),
  );
  writeFileSync(
    join(root, 'frontend', 'src', 'i18n', 'locales', 'en', 'common.json'),
    JSON.stringify({
      messageCode: { [`${game}.errExample`]: webText },
      ...(webLog ? { [`${game}.log.example`]: webLog } : {}),
      ...webEntries,
    }),
  );
  mkdirSync(join(root, 'internal', 'i18n', 'locales', 'ja'));
  mkdirSync(join(root, 'frontend', 'src', 'i18n', 'locales', 'ja'));
  writeFileSync(join(root, 'internal', 'i18n', 'locales', 'ja', 'example.json'), JSON.stringify({}));
  writeFileSync(
    join(root, 'frontend', 'src', 'i18n', 'locales', 'ja', 'common.json'),
    JSON.stringify({ messageCode: {} }),
  );
  return root;
}

function check(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8', cwd: process.cwd() });
}

describe('check-locale-text-agreement', () => {
  it('returns zero mismatches when shared text agrees', () => {
    const result = check(fixture('Same text.', 'Same text.', { cuiLog: 'logged', webLog: 'logged' }));
    expect(result.status).toBe(0);
    expect(`${result.stdout}${result.stderr}`).not.toContain('mismatch');
    expect(result.stdout).toContain('1 shared error messages');
    expect(result.stdout).toContain('1 shared log messages');
  });

  it('rejects a log message missing from the Web locale', () => {
    const result = check(fixture('Same text.', 'Same text.', { cuiLog: 'logged' }));
    expect(result.status).toBe(1);
    expect(`${result.stdout}${result.stderr}`).toContain('en example.log.example');
  });

  it('rejects a log message whose text differs', () => {
    const result = check(fixture('Same text.', 'Same text.', { cuiLog: 'CUI log', webLog: 'Web log' }));
    expect(result.status).toBe(1);
    expect(`${result.stdout}${result.stderr}`).toContain('CUI: CUI log');
    expect(`${result.stdout}${result.stderr}`).toContain('Web: Web log');
  });

  it('returns one mismatch and reports both texts when they disagree', () => {
    const result = check(fixture('CUI text.', 'Web text.'));
    const output = `${result.stdout}${result.stderr}`;
    expect(result.status).toBe(1);
    expect(output).toContain('1 mismatch');
    expect(output).toContain('en example.errExample');
    expect(output).toContain('CUI: CUI text.');
    expect(output).toContain('Web: Web text.');
  });

  it('compares a copied non-log key when the Web locale has it', () => {
    const result = check(
      fixture('Same text.', 'Same text.', {
        game: 'nap',
        cuiEntries: { 'bid.two': 'Bid two.' },
        webEntries: { 'nap.bid.two': 'Bid two.' },
      }),
    );
    expect(result.status).toBe(0);
    expect(result.stdout).toContain('1 copied keys compared');
  });

  it('rejects a copied non-log key whose text differs', () => {
    const result = check(
      fixture('Same text.', 'Same text.', {
        game: 'nap',
        cuiEntries: { 'bid.two': 'CUI bid.' },
        webEntries: { 'nap.bid.two': 'Web bid.' },
      }),
    );
    const output = `${result.stdout}${result.stderr}`;
    expect(result.status).toBe(1);
    expect(output).toContain('nap.bid.two');
  });

  it('does not compare an intentionally different key that is not copied', () => {
    const result = check(
      fixture('Same text.', 'Same text.', {
        cuiEntries: { helpBid: 'CUI help.' },
      }),
    );
    expect(result.status).toBe(0);
  });
});
