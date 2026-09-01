import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so resolve from cwd (the vitest root is
// `frontend/`) and assert the script is there — a wrong cwd would otherwise turn every case
// below into a spawn of a missing file rather than a visible failure.
const GUARD = join(process.cwd(), 'scripts', 'check-minchiate-last-trick.mjs');
if (!existsSync(GUARD)) throw new Error(`check-minchiate-last-trick.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** Build a fixture root holding just the two files the guard reads. */
function fixture(tsLine, goLine) {
  const dir = mkdtempSync(join(tmpdir(), 'minchiate-lt-'));
  dirs.push(dir);
  const pages = join(dir, 'frontend', 'src', 'pages');
  const presenter = join(dir, 'internal', 'domain');
  mkdirSync(pages, { recursive: true });
  mkdirSync(presenter, { recursive: true });
  writeFileSync(join(pages, 'MinchiatePage.tsx'), `${tsLine}\n`);
  writeFileSync(join(presenter, 'Minchiate.go'), `${goLine}\n`);
  return dir;
}

const run = (dir) => spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });

describe('check-minchiate-last-trick', () => {
  it('passes when both surfaces agree', () => {
    const r = run(fixture('const MINCHIATE_LAST_TRICK_BONUS = 3;', 'const MinchiateLastTrickBonus = 3'));
    expect(r.stdout).toContain('minchiate-last-trick: OK');
    expect(r.status).toBe(0);
  });

  // **A guard that only ever passes proves nothing.**
  it('fails when the two bonuss drift apart', () => {
    const r = run(fixture('const MINCHIATE_LAST_TRICK_BONUS = 4;', 'const MinchiateLastTrickBonus = 3'));
    expect(r.stderr).toContain('disagree');
    expect(r.status).toBe(1);
  });

  it('fails when a literal is renamed away', () => {
    const r = run(fixture('const LAST_TRICK_BONUS = 3;', 'const MinchiateLastTrickBonus = 3'));
    expect(r.stderr).toContain('could not find');
    expect(r.status).toBe(1);
  });
});
