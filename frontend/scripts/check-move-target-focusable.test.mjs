import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-move-target-focusable.mjs');
if (!existsSync(GUARD)) throw new Error(`check-move-target-focusable.mjs not found at ${GUARD}`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function check(source) {
  const dir = mkdtempSync(join(tmpdir(), 'check-move-target-focusable-'));
  dirs.push(dir);
  const page = join(dir, 'src', 'pages', 'example.tsx');
  mkdirSync(join(dir, 'src', 'pages'), { recursive: true });
  writeFileSync(page, source);
  const result = spawnSync('bun', [GUARD, join(dir, 'src')], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

function checkEmpty() {
  const dir = mkdtempSync(join(tmpdir(), 'check-move-target-focusable-empty-'));
  dirs.push(dir);
  mkdirSync(join(dir, 'src'), { recursive: true });
  const result = spawnSync('bun', [GUARD, join(dir, 'src')], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-move-target-focusable', () => {
  it('flags a selection term at the top level', () => {
    const result = check('<button disabled={!isPlaying || loading || !selectedSource} />;\n');
    expect(result.code).toBe(1);
    expect(result.out).toContain('pages/example.tsx:1: disabled={!isPlaying || loading || !selectedSource}');
    expect(result.out).toContain('aria-disabled');
  });

  it('flags a multi-line expression ending in a target selection term', () => {
    const result = check('<button disabled={\n  !isPlaying || loading ||\n  !isTarget\n} />;\n');
    expect(result.code).toBe(1);
    expect(result.out).toContain('disabled={!isPlaying || loading ||\n  !isTarget}');
  });

  it('allows aria-disabled selection guards', () => {
    const result = check('<button aria-disabled={!selectedSource || undefined} />;\n');
    expect(result.code).toBe(0);
    expect(result.out).toContain('move-target-focusable: OK');
  });

  it('allows selected null checks combined with loading', () => {
    const result = check('<button disabled={loading || selected === null} />;\n');
    expect(result.code).toBe(0);
    expect(result.out).toContain('move-target-focusable: OK');
  });

  it('allows a parenthesized buried-card condition', () => {
    const result = check('<button disabled={!isPlaying || loading || (!isTop && !selectedSource)} />;\n');
    expect(result.code).toBe(0);
    expect(result.out).toContain('move-target-focusable: OK');
  });

  it('allows unrelated selection-count conditions', () => {
    const result = check('<button disabled={loading || selectedCardIndices.length !== 1} />;\n');
    expect(result.code).toBe(0);
    expect(result.out).toContain('move-target-focusable: OK');
  });

  it('rejects an empty scan tree', () => {
    const result = checkEmpty();
    expect(result.code).toBe(1);
    expect(result.out).toContain('move-target-focusable: scanned 0 files');
  });
});
