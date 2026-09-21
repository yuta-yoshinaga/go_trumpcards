import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-action-log-params.mjs');
if (!existsSync(GUARD)) throw new Error(`check-action-log-params.mjs not found at ${GUARD}`);
const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(source, ja, en = ja) {
  const root = mkdtempSync(join(tmpdir(), 'action-log-params-'));
  dirs.push(root);
  mkdirSync(join(root, 'internal', 'domain'), { recursive: true });
  mkdirSync(join(root, 'internal', 'i18n', 'locales', 'ja'), { recursive: true });
  mkdirSync(join(root, 'internal', 'i18n', 'locales', 'en'), { recursive: true });
  writeFileSync(join(root, 'internal', 'domain', 'Fixture.go'), `package domain\n${source}\n`);
  writeFileSync(join(root, 'internal', 'i18n', 'locales', 'ja', 'fixture.json'), JSON.stringify(ja));
  writeFileSync(join(root, 'internal', 'i18n', 'locales', 'en', 'fixture.json'), JSON.stringify(en));
  return root;
}

function run(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8' });
}

describe('check-action-log-params', () => {
  it('accepts matching placeholders and params', () => {
    const r = run(
      fixture('func f() { appendLog(0, "play", "fixture.log.play", map[string]string{"name": "x"}, nil) }', {
        'log.play': '{{name}} plays',
      }),
    );
    expect(r.status).toBe(0);
  });

  it('accepts Key-suffixed params against the resolved placeholder name', () => {
    const r = run(
      fixture(
        'func f() { appendLog(0, "play", "fixture.log.play", map[string]string{"suitKey": "common.suit.spade"}, nil) }',
        {
          'log.play': 'trump {{suit}}',
        },
      ),
    );
    expect(r.status).toBe(0);
  });

  it('rejects a mismatched param key and prints the key name', () => {
    const r = run(
      fixture('func f() { appendLog(0, "play", "fixture.log.play", map[string]string{"cardStr": "x"}, nil) }', {
        'log.play': '{{card}} played',
      }),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('card');
    expect(r.stderr).toContain('cardStr');
  });

  it('rejects a code missing from locale', () => {
    const r = run(fixture('func f() { appendLog(0, "play", "fixture.log.play", nil, nil) }', { 'log.other': 'other' }));
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('fixture.log.play');
  });

  it('accepts nil params when the locale has no placeholders', () => {
    const r = run(fixture('func f() { appendLog(0, "play", "fixture.log.play", nil, nil) }', { 'log.play': 'played' }));
    expect(r.status).toBe(0);
  });

  it('finds code literals passed through an addLog wrapper', () => {
    const r = run(
      fixture(
        `func f() {
  addLog("fixture.log.play", map[string]string{"name": "x"})
}
func addLog(code string, params map[string]string) {
  appendLogCode(code, params)
}`,
        { 'log.play': '{{name}} plays' },
      ),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('checked 1 codes');
  });

  it('reports code literals whose params are hidden behind a variable', () => {
    const r = run(
      fixture(
        `func f(params map[string]string) {
  addLog("fixture.log.play", params)
}`,
        { 'log.play': '{{name}} plays' },
      ),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 codes skipped because params could not be read');
  });
});
