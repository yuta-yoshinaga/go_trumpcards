import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-domain-error-locale.mjs');
if (!existsSync(GUARD)) throw new Error(`check-domain-error-locale.mjs not found at ${GUARD}`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(...files) {
  const root = mkdtempSync(join(tmpdir(), 'domain-error-locale-'));
  dirs.push(root);
  const domain = join(root, 'internal', 'domain');
  mkdirSync(domain, { recursive: true });
  for (const [name, source] of files) writeFileSync(join(domain, name), source);
  return root;
}

function run(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8' });
}

describe('check-domain-error-locale', () => {
  it('counts Japanese literals in the second argument only', () => {
    const r = run(
      fixture(
        ['Japanese.go', 'package domain\nfunc f() error { return NewDomainError(ErrX, "日本語") }\n'],
        ['Ascii.go', 'package domain\nfunc f() error { return NewDomainError(ErrX, "plain English") }\n'],
        ['Code.go', 'package domain\nfunc f() error { return NewDomainErrorCode(ErrX, "日本語") }\n'],
      ),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 Japanese literals in NewDomainError across 1 file(s)');
  });

  it('counts a multiline call, and does not count Japanese in the sentinel', () => {
    const r = run(
      fixture([
        'Multiline.go',
        'package domain\nfunc f() error { return NewDomainError(\n Err日本語,\n "エラー",\n) }\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 Japanese literals');
  });

  it('counts Japanese in the first fmt.Sprintf argument', () => {
    const r = run(
      fixture([
        'Formatted.go',
        'package domain\nfunc f() error { return NewDomainError(ErrX, fmt.Sprintf("%d 枚必要です", n)) }\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 Japanese literals');
  });

  it('counts a fmt.Sprintf argument when it starts on the next line', () => {
    const r = run(
      fixture([
        'MultilineFormatted.go',
        'package domain\nfunc f() error { return NewDomainError(\n ErrX,\n fmt.Sprintf("あと %d 枚", n),\n) }\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('1 Japanese literals');
  });

  it('does not count Japanese in later fmt.Sprintf arguments', () => {
    const r = run(
      fixture([
        'FormatArgument.go',
        'package domain\nfunc f() error { return NewDomainError(ErrX, fmt.Sprintf("need %d cards", "日本語")) }\n',
      ]),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('0 Japanese literals');
  });

  it('does not count NewDomainErrorCode', () => {
    const r = run(
      fixture(['CodeOnly.go', 'package domain\nfunc f() error { return NewDomainErrorCode(ErrX, "日本語") }\n']),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('0 Japanese literals');
  });

  it('fails when the ceiling is exceeded', () => {
    const calls = Array.from(
      { length: 793 },
      (_, i) => `func f${i}() error { return NewDomainError(ErrX, "日本語") }`,
    ).join('\n');
    const r = run(fixture(['TooMany.go', `package domain\n${calls}\n`]));
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('exceeds ceiling');
  });
});
