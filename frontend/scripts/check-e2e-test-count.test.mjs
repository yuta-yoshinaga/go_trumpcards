import { spawnSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const SCRIPT = join(process.cwd(), 'scripts/check-e2e-test-count.mjs');
const dirs = [];
afterAll(() => {
  for (const dir of dirs) rmSync(dir, { recursive: true, force: true });
});

function run(output, limit = '500') {
  const dir = mkdtempSync(join(tmpdir(), 'e2e-count-'));
  dirs.push(dir);
  const bin = join(dir, 'bin');
  const data = join(dir, 'listing.txt');
  const baseline = join(dir, 'baseline.json');
  const args = join(dir, 'args.txt');
  const script = join(bin, 'bunx');
  mkdirSync(bin);
  const quotedData = data.replaceAll("'", "'\\''");
  const quotedArgs = args.replaceAll("'", "'\\''");
  writeFileSync(data, output);
  writeFileSync(baseline, JSON.stringify({ max: Number(limit), reason: 'test fixture' }));
  writeFileSync(script, `#!/bin/sh\nprintf '%s' "$*" > '${quotedArgs}'\ncat '${quotedData}'\n`);
  chmodSync(script, 0o755);
  const result = spawnSync('bun', [SCRIPT], {
    encoding: 'utf8',
    env: {
      ...process.env,
      PATH: `${bin}:${process.env.PATH}`,
      PLAYWRIGHT_LIST_GUARD_BASELINE: baseline,
    },
  });
  return {
    status: result.status,
    out: `${result.stdout}${result.stderr}`,
    args: existsSync(args) ? readFileSync(args, 'utf8') : '',
  };
}

describe('check-e2e-test-count', () => {
  it('passes at the baseline and invokes Playwright list mode only', () => {
    const { status, out, args } = run('Total: 4000 tests in 12 files\n', '4000');
    expect(status).toBe(0);
    expect(out).toContain('4000');
    expect(args).toBe('playwright test --list');
  });

  it('fails above the baseline', () => {
    const { status, out } = run('Total: 4001 tests in 12 files\n', '4000');
    expect(status).toBe(1);
    expect(out).toContain('4001');
  });

  it('fails when the listing has no tests', () => {
    const { status, out } = run('No tests found\n');
    expect(status).toBe(1);
    expect(out).toContain('at least');
  });
});
