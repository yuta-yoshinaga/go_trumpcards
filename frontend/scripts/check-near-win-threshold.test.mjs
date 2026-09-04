import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so resolve from cwd (the vitest root is
// `frontend/`) and assert the script is there — a wrong cwd would otherwise turn every case
// below into a spawn of a missing file rather than a visible failure.
const GUARD = join(process.cwd(), 'scripts', 'check-near-win-threshold.mjs');
if (!existsSync(GUARD)) throw new Error(`check-near-win-threshold.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** Build a fixture root holding just the two files the guard reads. */
function fixture(tsLine, goLine) {
  const dir = mkdtempSync(join(tmpdir(), 'near-win-'));
  dirs.push(dir);
  const pages = join(dir, 'frontend', 'src', 'pages');
  const presenter = join(dir, 'internal', 'adapter', 'presenter');
  mkdirSync(pages, { recursive: true });
  mkdirSync(presenter, { recursive: true });
  writeFileSync(join(pages, 'TysiacPage.tsx'), `${tsLine}\n`);
  writeFileSync(join(presenter, 'TysiacCuiPresenter.go'), `${goLine}\n`);
  return dir;
}

const run = (dir) => spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });

describe('check-near-win-threshold', () => {
  it('passes when both surfaces agree', () => {
    const r = run(fixture('const NEAR_WIN_RATIO = 0.8;', 'const tysiacNearWinRatio = 0.8'));
    expect(r.stdout).toContain('near-win-threshold: OK');
    expect(r.status).toBe(0);
  });

  // **A guard that only ever passes proves nothing.**
  it('fails when the two thresholds drift apart', () => {
    const r = run(fixture('const NEAR_WIN_RATIO = 0.9;', 'const tysiacNearWinRatio = 0.8'));
    expect(r.stderr).toContain('disagree');
    expect(r.status).toBe(1);
  });

  it('fails when a literal is renamed away', () => {
    const r = run(fixture('const SOMETHING_ELSE = 0.8;', 'const tysiacNearWinRatio = 0.8'));
    expect(r.stderr).toContain('could not find');
    expect(r.status).toBe(1);
  });
});
