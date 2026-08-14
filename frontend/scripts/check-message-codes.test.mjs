import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-message-codes.mjs');
if (!existsSync(GUARD)) throw new Error(`check-message-codes.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/**
 * Build a presenter dir + locale dir pair.
 *
 * `codes` are the messageCode strings the fake locale files define; the presenter source
 * is supplied verbatim so each case can choose the emission form under test.
 */
function fixture(presenterSrc, codes) {
  const dir = mkdtempSync(join(tmpdir(), 'message-codes-'));
  dirs.push(dir);
  const presenters = join(dir, 'presenter');
  mkdirSync(presenters);
  writeFileSync(join(presenters, 'DemoWebPresenter.go'), presenterSrc);
  for (const lang of ['ja', 'en']) {
    mkdirSync(join(dir, 'locales', lang), { recursive: true });
    const messageCode = Object.fromEntries(codes.map((c) => [c, `${lang}:${c}`]));
    writeFileSync(join(dir, 'locales', lang, 'common.json'), JSON.stringify({ messageCode }));
  }
  return { presenters, locales: join(dir, 'locales') };
}

function check(presenterSrc, codes) {
  const { presenters, locales } = fixture(presenterSrc, codes);
  const r = spawnSync(process.execPath, [GUARD, presenters, locales], { encoding: 'utf8' });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

const RETURN_FORM = `func (p *DemoWebPresenter) buildMessage() (string, string, map[string]string) {
	return "", "demo.returned", nil
}
`;

/** The form the guard was blind to: the code is assigned to a field, never returned. */
const ASSIGN_FORM = `func (p *DemoWebPresenter) Output() string {
	resObj.Message = "ゲームクリア！"
	resObj.MessageCode = "demo.assigned"
	return marshalOrError(resObj)
}
`;

describe('check-message-codes', () => {
  it('accepts a returned code that both locales translate', () => {
    const r = check(RETURN_FORM, ['demo.returned']);
    expect(r.code).toBe(0);
    expect(r.out).toContain('OK');
  });

  it('rejects a returned code with no translation', () => {
    const r = check(RETURN_FORM, []);
    expect(r.code).toBe(1);
    expect(r.out).toContain('demo.returned');
  });

  // **代入形式こそがこのガードの盲点だった。** 拾えていなければ「0 件」となり、
  // 翻訳が 1 つも無くても OK と報告してしまう。
  it('sees the assignment form and rejects it when untranslated', () => {
    const r = check(ASSIGN_FORM, []);
    expect(r.code).toBe(1);
    expect(r.out).toContain('demo.assigned');
  });

  it('accepts the assignment form once both locales translate it', () => {
    const r = check(ASSIGN_FORM, ['demo.assigned']);
    expect(r.code).toBe(0);
    expect(r.out).toContain('OK');
  });

  // **代入形式は「空箱」か「生リテラル」かをこの行から読めない。** Message は
  // 別の文で、そもそも書かない分岐もある (Colorado の playing/stalemate)。
  // 近さで推測せず、専用のバケットで報告する。
  it('reports the assignment form in its own bucket', () => {
    const r = check(ASSIGN_FORM, []);
    expect(r.code).toBe(1);
    expect(r.out).toContain('resObj.MessageCode =');
    expect(r.out).toContain('or an empty box if it assigned none');
  });

  it('rejects a code translated in only one locale', () => {
    const { presenters, locales } = fixture(ASSIGN_FORM, ['demo.assigned']);
    writeFileSync(join(locales, 'en', 'common.json'), JSON.stringify({ messageCode: {} }));
    const r = spawnSync(process.execPath, [GUARD, presenters, locales], { encoding: 'utf8' });
    expect(r.status).toBe(1);
    expect(`${r.stdout}${r.stderr}`).toContain('[en]');
  });
});
