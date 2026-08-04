import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// Vitest serves test modules under a non-file URL, so `import.meta.url` cannot be converted
// with fileURLToPath here. The vitest root is `frontend/`, so resolve from cwd — and assert
// the script is actually there, because a wrong cwd would otherwise turn every case below
// into a spawn of a missing file rather than a visible failure.
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-guard-floors.mjs');
if (!existsSync(GUARD)) throw new Error(`check-guard-floors.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** A guard that reports one count and floors it — the shape the checker should accept. */
const SOUND = `import { assertFloor } from './lib/floor.mjs';
const items = await walk();
assertFloor('demo', items.length, 100, 'items');
console.log(\`demo: OK (\${items.length} items scanned).\`);
`;

/**
 * Build a fixture directory of guard scripts.
 *
 * The default is eight sound guards because the checker floors its own walk at eight — a
 * fixture of two would fail for the wrong reason and prove nothing about the rule under test.
 */
function fixture(extra = {}, sound = 8) {
  const dir = mkdtempSync(join(tmpdir(), 'guard-floors-'));
  dirs.push(dir);
  for (let i = 0; i < sound; i += 1) writeFileSync(join(dir, `check-sound-${i}.mjs`), SOUND);
  for (const [name, src] of Object.entries(extra)) writeFileSync(join(dir, name), src);
  return dir;
}

function check(dir) {
  const r = spawnSync(process.execPath, [GUARD, dir], { encoding: 'utf8' });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-guard-floors', () => {
  it('accepts guards that floor every count they report', () => {
    const r = check(fixture());
    expect(r.code).toBe(0);
    expect(r.out).toContain('guard-floors: OK');
  });

  it('rejects a guard that reports a count with no floor under it', () => {
    const r = check(
      fixture({
        // biome-ignore lint/suspicious/noTemplateCurlyInString: fixture source text, not an interpolation
        'check-unfloored.mjs': 'const items = await walk();\nconsole.log(`demo: OK (${items.length} scanned).`);\n',
      }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('check-unfloored.mjs');
    expect(r.out).toContain('asserts 0 floor(s)');
  });

  it('rejects a guard with more reported counts than floors', () => {
    const r = check(
      fixture({
        'check-partial.mjs': `import { assertFloor } from './lib/floor.mjs';
assertFloor('partial', files.length, 50, 'files');
console.log(\`a: OK (\${files.length} files).\`);
console.log(\`b: OK (\${hooks.length} hooks).\`);
`,
      }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('check-partial.mjs');
    expect(r.out).toContain('2 discovered count(s) but asserts 1 floor(s)');
  });

  it('rejects a floor of zero, which holds for every possible input', () => {
    const r = check(
      fixture({
        'check-zero.mjs': `import { assertFloor } from './lib/floor.mjs';
assertFloor('zero', items.length, 0, 'items');
console.log(\`zero: OK (\${items.length} items).\`);
`,
      }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('floor is 0');
  });

  it('rejects a floor that cannot be read off the source', () => {
    const r = check(
      fixture({
        'check-computed.mjs': `import { assertFloor } from './lib/floor.mjs';
assertFloor('computed', items.length, items.length, 'items');
console.log(\`computed: OK (\${items.length} items).\`);
`,
      }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('is not a literal or a same-file const');
  });

  it('resolves a floor written as a same-file const', () => {
    const r = check(
      fixture({
        'check-const.mjs': `import { assertFloor } from './lib/floor.mjs';
const MIN_ITEMS = 300;
assertFloor('constant', items.length, MIN_ITEMS, 'items');
console.log(\`constant: OK (\${items.length} items).\`);
`,
      }),
    );
    expect(r.code).toBe(0);
  });

  it('accepts a guard whose success line reports no count', () => {
    const r = check(
      fixture({
        'check-boolean.mjs': "console.log('boolean: OK (index.css has the universal block).');\n",
      }),
    );
    expect(r.code).toBe(0);
  });

  it('ignores a guard’s own test file, whose fixtures are deliberately broken', () => {
    const r = check(
      fixture({
        // biome-ignore lint/suspicious/noTemplateCurlyInString: fixture source text, not an interpolation
        'check-sound-0.test.mjs': 'console.log(`fixture: OK (${n} items)`); // no floor, on purpose\n',
      }),
    );
    expect(r.code).toBe(0);
  });

  it('fails on a directory holding too few guards to be the real one', () => {
    const r = check(fixture({}, 3));
    expect(r.code).toBe(1);
    expect(r.out).toContain('only 3 guard scripts');
  });

  it('passes against the repository’s own guards', () => {
    // The live control. Every case above uses synthetic input; without this one the suite
    // would stay green after a change that rejects every real guard in the directory.
    const r = check(SCRIPTS);
    expect(r.code).toBe(0);
    expect(r.out).toMatch(/guard-floors: OK \(\d+ guard scripts/);
  });
});
