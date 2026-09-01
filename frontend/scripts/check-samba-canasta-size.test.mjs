import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so resolve from cwd (the vitest root is
// `frontend/`) and assert the script is there — a wrong cwd would otherwise turn every case
// below into a spawn of a missing file rather than a visible failure.
const GUARD = join(process.cwd(), 'scripts', 'check-samba-canasta-size.mjs');
if (!existsSync(GUARD)) throw new Error(`check-samba-canasta-size.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** Build a fixture root holding just the two files the guard reads. */
function fixture(tsLine, goLine) {
  const dir = mkdtempSync(join(tmpdir(), 'samba-size-'));
  dirs.push(dir);
  const pages = join(dir, 'frontend', 'src', 'pages');
  const domain = join(dir, 'internal', 'domain');
  mkdirSync(pages, { recursive: true });
  mkdirSync(domain, { recursive: true });
  writeFileSync(join(pages, 'SambaPage.tsx'), `${tsLine}\n`);
  writeFileSync(join(domain, 'SambaPlayer.go'), `${goLine}\n`);
  return dir;
}

const run = (dir) => spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });

describe('check-samba-canasta-size', () => {
  it('passes when both surfaces agree', () => {
    const r = run(fixture('const SAMBA_CANASTA_SIZE = 7;', 'const SambaCanastaSize = 7'));
    expect(r.stdout).toContain('samba-canasta-size: OK');
    expect(r.status).toBe(0);
  });

  // **A guard that only ever passes proves nothing.**
  it('fails when the two sizes drift apart', () => {
    const r = run(fixture('const SAMBA_CANASTA_SIZE = 7;', 'const SambaCanastaSize = 8'));
    expect(r.stderr).toContain('disagree');
    expect(r.status).not.toBe(0);
  });

  // A rename is the silent way a pair stops being a pair.
  it('fails when a side is renamed out of reach', () => {
    const r = run(fixture('const CANASTA_SIZE = 7;', 'const SambaCanastaSize = 7'));
    expect(r.stderr).toContain('could not find the size');
    expect(r.status).not.toBe(0);
  });
});
