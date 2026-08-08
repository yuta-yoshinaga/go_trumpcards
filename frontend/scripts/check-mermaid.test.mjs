import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so `import.meta.url` cannot be converted
// with fileURLToPath here. The vitest root is `frontend/`, so resolve from cwd — and assert
// the script is actually there, because a wrong cwd would otherwise turn every case below
// into a spawn of a missing file rather than a visible failure.
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-mermaid.mjs');
if (!existsSync(GUARD)) throw new Error(`check-mermaid.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

const VALID = `# Doc

\`\`\`mermaid
flowchart TD
    A[Start] --> B[End]
\`\`\`
`;

/** Unquoted parentheses inside a node label — the exact break found in skat.md. */
const BROKEN = `# Doc

\`\`\`mermaid
flowchart TD
    A[Start] --> B[bid (b 0/1)]
\`\`\`
`;

/** A fenced block that is not mermaid; the walk must ignore it entirely. */
const NOT_MERMAID = `# Doc

\`\`\`text
flowchart TD
    this is not a diagram
\`\`\`
`;

function fixture(files) {
  const dir = mkdtempSync(join(tmpdir(), 'check-mermaid-'));
  dirs.push(dir);
  for (const [name, src] of Object.entries(files)) writeFileSync(join(dir, name), src);
  return dir;
}

function check(dir) {
  const r = spawnSync(process.execPath, [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-mermaid', () => {
  // The positive control. A guard that only ever fires proves nothing about the
  // day it stays silent, so assert it passes correct input before asserting it
  // rejects anything.
  it('accepts diagrams that parse', () => {
    const r = check(fixture({ 'a.md': VALID, 'b.md': VALID }));
    expect(r.code).toBe(0);
    expect(r.out).toContain('mermaid: OK');
    expect(r.out).toContain('2 blocks across 2 files');
  });

  it('rejects a diagram that does not parse, naming the file and line', () => {
    const r = check(fixture({ 'good.md': VALID, 'bad.md': BROKEN }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('1 block(s) failed to parse');
    expect(r.out).toContain('bad.md:3');
    // The passing file must not be reported as a failure.
    expect(r.out).not.toContain('good.md:');
  });

  it('ignores fenced blocks that are not mermaid', () => {
    const r = check(fixture({ 'a.md': VALID, 'b.md': NOT_MERMAID }));
    expect(r.code).toBe(0);
    expect(r.out).toContain('1 blocks across 1 files');
  });

  it('reports every failure, not just the first', () => {
    const r = check(fixture({ 'x.md': BROKEN, 'y.md': BROKEN }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('2 block(s) failed to parse');
    expect(r.out).toContain('x.md:3');
    expect(r.out).toContain('y.md:3');
  });

  it('descends into subdirectories', () => {
    const dir = fixture({ 'top.md': VALID });
    mkdirSync(join(dir, 'docs', 'design'), { recursive: true });
    writeFileSync(join(dir, 'docs', 'design', 'deep.md'), BROKEN);
    const r = check(dir);
    expect(r.code).toBe(1);
    expect(r.out).toContain('docs/design/deep.md:3');
  });

  it('skips node_modules', () => {
    const dir = fixture({ 'top.md': VALID });
    mkdirSync(join(dir, 'node_modules', 'pkg'), { recursive: true });
    writeFileSync(join(dir, 'node_modules', 'pkg', 'readme.md'), BROKEN);
    const r = check(dir);
    expect(r.code).toBe(0);
    expect(r.out).toContain('1 blocks across 1 files');
  });
});
