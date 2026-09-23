import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-error-code-presenters.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(presenterSource) {
  const root = mkdtempSync(join(tmpdir(), 'error-code-presenters-'));
  dirs.push(root);
  mkdirSync(join(root, 'internal', 'domain'), { recursive: true });
  mkdirSync(join(root, 'internal', 'adapter', 'presenter'), { recursive: true });
  writeFileSync(join(root, 'internal', 'domain', 'Example.go'), 'package domain\n\nvar _ = NewDomainErrorCode\n');
  writeFileSync(join(root, 'internal', 'adapter', 'presenter', 'ExampleWebPresenter.go'), presenterSource);
  return root;
}

function check(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8', cwd: process.cwd() });
}

describe('check-error-code-presenters', () => {
  it('accepts a presenter that uses ErrorMessageCode', () => {
    const result = check(fixture('package presenter\n\nvar _ = ErrorMessageCode\nvar _ = lastErr.Error()\n'));
    expect(result.status).toBe(0);
  });

  it('rejects a presenter that exposes lastErr.Error() without a safe path', () => {
    const result = check(fixture('package presenter\n\nvar _ = lastErr.Error()\n'));
    const output = `${result.stdout}${result.stderr}`;
    expect(result.status).toBe(1);
    expect(output).toContain('ExampleWebPresenter.go');
  });

  it('accepts a presenter that uses cuiErrorBlock', () => {
    const result = check(fixture('package presenter\n\nvar _ = cuiErrorBlock\nvar _ = lastErr.Error()\n'));
    expect(result.status).toBe(0);
  });
});
