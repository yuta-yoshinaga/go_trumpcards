#!/usr/bin/env bun
// Keep the full E2E suite within the CI shard budget by pinning its collected test count.

import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const SCRIPT_DIR = fileURLToPath(new URL('.', import.meta.url));
// PLAYWRIGHT_LIST_GUARD_BASELINE is the test seam for supplying a fixture baseline.
const BASELINE_PATH = process.env.PLAYWRIGHT_LIST_GUARD_BASELINE ?? `${SCRIPT_DIR}e2e-test-count-baseline.json`;
const MIN_TESTS = 1550;

let baseline;
try {
  baseline = JSON.parse(await readFile(BASELINE_PATH, 'utf8'));
} catch {
  console.error(`e2e-test-count: could not read or parse baseline JSON at ${BASELINE_PATH}`);
  process.exit(1);
}
if (!Number.isInteger(baseline.max) || baseline.max <= 0) {
  console.error(`e2e-test-count: invalid max in ${BASELINE_PATH}`);
  process.exit(1);
}

const result = Bun.spawnSync(['bunx', 'playwright', 'test', '--list'], {
  cwd: fileURLToPath(new URL('..', import.meta.url)),
});
const stdout = new TextDecoder().decode(result.stdout);
const stderr = new TextDecoder().decode(result.stderr);
if (result.exitCode !== 0) {
  console.error('e2e-test-count: could not list Playwright tests.');
  if (stderr) console.error(stderr.trim());
  process.exit(1);
}

const match = `${stdout}\n${stderr}`.match(/^Total:\s+(\d+) tests?\b/m);
const count = match ? Number(match[1]) : 0;
assertFloor('e2e-test-count', count, MIN_TESTS, 'Playwright E2E tests');

if (count > baseline.max) {
  console.error(
    `e2e-test-count: ${count} tests exceed the ratchet limit of ${baseline.max}. ` +
      'If this increase is intentional, update scripts/e2e-test-count-baseline.json and explain why in its reason.',
  );
  process.exit(1);
}

console.log(`e2e-test-count: OK (${count} tests; limit ${baseline.max}).`);
