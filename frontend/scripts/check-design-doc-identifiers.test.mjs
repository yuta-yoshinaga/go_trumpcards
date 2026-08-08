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
const GUARD = join(SCRIPTS, 'check-design-doc-identifiers.mjs');
if (!existsSync(GUARD)) throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** A design doc whose diagram names the given classes. */
function doc(...classes) {
  const body = classes.map((c) => `    class ${c} {\n        +thing()\n    }`).join('\n');
  return ['# Design', '', '```mermaid', 'classDiagram', body, '```', ''].join('\n');
}

function fixture({ sources = {}, classes = [] }) {
  const dir = mkdtempSync(join(tmpdir(), 'design-doc-ids-'));
  dirs.push(dir);
  const src = join(dir, 'src');
  mkdirSync(src, { recursive: true });
  for (const [name, body] of Object.entries(sources)) {
    const full = join(src, name);
    mkdirSync(join(full, '..'), { recursive: true });
    writeFileSync(full, body);
  }
  const docFile = join(dir, 'design.md');
  writeFileSync(docFile, doc(...classes));
  return { src, docFile };
}

function check({ src, docFile }) {
  const r = spawnSync(process.execPath, [GUARD, src, docFile], { encoding: 'utf8', cwd: process.cwd() });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-design-doc-identifiers', () => {
  // Positive control first: a guard that only ever fires proves nothing about
  // the day it stays quiet.
  it('accepts a class backed by a named export', () => {
    const r = check(fixture({ sources: { 'a.ts': 'export function Widget() {}' }, classes: ['Widget'] }));
    expect(r.code).toBe(0);
    expect(r.out).toContain('design-doc-identifiers: OK');
  });

  it('rejects a class with no counterpart', () => {
    const r = check(fixture({ sources: { 'a.ts': 'export function Widget() {}' }, classes: ['Ghost'] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('Ghost');
  });

  it('accepts a class that names a module rather than an export', () => {
    // `gameApi`, `urlMoodCodec` and friends are drawn as module nodes.
    const r = check(fixture({ sources: { 'gameApi.ts': 'export const x = 1;' }, classes: ['gameApi'] }));
    expect(r.code).toBe(0);
  });

  it('accepts a class exported via `export default <ident>`', () => {
    const r = check(fixture({ sources: { 'a.ts': 'const i18n = {};\nexport default i18n;' }, classes: ['i18n'] }));
    expect(r.code).toBe(0);
  });

  it('accepts a class exported through an export list, using the renamed binding', () => {
    const r = check(fixture({ sources: { 'a.ts': 'const a = 1;\nexport { a as Renamed };' }, classes: ['Renamed'] }));
    expect(r.code).toBe(0);
  });

  it('accepts the documented placeholders', () => {
    const r = check(fixture({ sources: { 'a.ts': 'export const x = 1;' }, classes: ['GamePage', 'DurakPhase'] }));
    expect(r.code).toBe(0);
  });

  it('does not credit exports that live only in test files', () => {
    const r = check(
      fixture({ sources: { 'a.test.ts': 'export function OnlyInTests() {}' }, classes: ['OnlyInTests'] }),
    );
    expect(r.code).toBe(1);
    expect(r.out).toContain('OnlyInTests');
  });

  it('reports every unresolved class, not just the first', () => {
    const r = check(fixture({ sources: { 'a.ts': 'export const x = 1;' }, classes: ['GhostOne', 'GhostTwo'] }));
    expect(r.code).toBe(1);
    expect(r.out).toContain('GhostOne');
    expect(r.out).toContain('GhostTwo');
  });
});
