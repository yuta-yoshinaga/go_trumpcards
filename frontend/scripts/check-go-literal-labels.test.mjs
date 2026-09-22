import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-go-literal-labels.mjs');
if (!existsSync(GUARD)) throw new Error(`check-go-literal-labels.mjs not found at ${GUARD}`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function fixture(...files) {
  const root = mkdtempSync(join(tmpdir(), 'go-literal-labels-'));
  dirs.push(root);
  const presenter = join(root, 'internal', 'adapter', 'presenter');
  mkdirSync(presenter, { recursive: true });
  for (const [name, source] of files) writeFileSync(join(presenter, name), source);
  return root;
}

function run(root) {
  return spawnSync('bun', [GUARD, root], { encoding: 'utf8' });
}

describe('check-go-literal-labels', () => {
  it('detects an English label followed by a format verb', () => {
    const r = run(fixture(['Presenter.go', 'package presenter\nfunc f() { _ = fmt.Sprintf("score=%d", 1) }\n']));
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('score=%d');
  });

  it('does not detect a colon without a format verb', () => {
    const r = run(fixture(['Presenter.go', 'package presenter\nvar _ = "name: value"\n']));
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('0 English labels');
  });

  it('does not scan _test.go files', () => {
    const r = run(fixture(['Presenter_test.go', 'package presenter\nvar _ = "score=%d"\n']));
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('0 English labels');
  });

  it('allows a translated i18n.Tf format literal', () => {
    const r = run(
      fixture(['Presenter.go', 'package presenter\nfunc f() { _ = i18n.Tf("game.playerLine", "score", 1) }\n']),
    );
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('0 English labels');
  });
});
