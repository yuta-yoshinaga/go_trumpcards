import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-suit-symbols.mjs');
if (!existsSync(GUARD)) throw new Error(`check-suit-symbols.mjs not found at ${GUARD}`);

const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function check(source) {
  const dir = mkdtempSync(join(tmpdir(), 'check-suit-symbols-'));
  dirs.push(dir);
  writeFileSync(join(dir, 'example.ts'), source);
  const result = spawnSync('bun', [GUARD, dir], { encoding: 'utf8', cwd: process.cwd() });
  return { code: result.status, out: `${result.stdout}${result.stderr}` };
}

describe('check-suit-symbols', () => {
  it('accepts source without a local table', () => {
    const result = check('export const suit = 1;\n');
    expect(result.code).toBe(0);
    expect(result.out).toContain('suit-symbols: OK');
  });

  it('rejects local tables and reports file and line', () => {
    const result = check('const other = 1;\nlet SUIT_SYMBOLS = [];\n');
    expect(result.code).toBe(1);
    expect(result.out).toContain('example.ts:2');
    expect(result.out).toContain('suitSymbolAt(suit, fallback) / suitSymbol(design) (#8022)');
  });
});
