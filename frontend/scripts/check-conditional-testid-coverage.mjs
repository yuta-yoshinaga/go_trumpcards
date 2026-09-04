#!/usr/bin/env bun
/**
 * Guard that data-testids inside conditional rendering blocks are referenced by tests or E2E.
 *
 * A data-testid nested inside a conditional rendering block (`{cond && (...)}` or
 * `{cond ? (...) : ...}`) will never be rendered unless test fixtures satisfy that condition.
 * For example, CourchevelHiLoPage's `cv-preflop-exposed-note` went completely unrendered
 * because no fixture satisfied the condition, yet 124 tests stayed green (issue #7074).
 *
 * This guard ensures every statically declared data-testid within conditional JSX blocks
 * in page components is referenced in at least one test (`src/**\/*.test.ts(x)`) or E2E spec (`e2e/**\/*.ts`).
 */

import { readdir, readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

// Fixture mode is opt-in via `--root <dir>` and is the only mode that skips the floors.
// A bare positional argument must NOT disable them: a floor that a stray argv can switch
// off is a floor that reports coverage it never checked.
const rootFlag = process.argv.indexOf('--root');
const SCANNING_REPO = rootFlag === -1;
const FRONTEND = SCANNING_REPO ? fileURLToPath(new URL('..', import.meta.url)) : resolve(process.argv[rootFlag + 1]);

const PAGES_DIR = join(FRONTEND, 'src/pages');
const COMPONENTS_DIR = join(FRONTEND, 'src/components');
const SRC_DIR = join(FRONTEND, 'src');
const E2E_DIR = join(FRONTEND, 'e2e');

/**
 * Finds conditional data-testid attributes in JSX source text.
 * Walks backwards from each data-testid line to determine if it is inside
 * a conditional rendering expression (`{...&& (` or `{...? (` or `^{...&& <`).
 *
 * @param {string} content - JSX source text.
 * @returns {Array<{ tid: string, line: number, testidLine: number, ctx: string }>}
 */
export function extractConditionalTestids(content) {
  const lines = content.split('\n');
  const out = [];
  for (let i = 0; i < lines.length; i++) {
    const ln = lines[i];
    const m = ln.match(/data-testid=["']([^"']+)["']/);
    if (!m) continue;
    const tid = m[1];
    let ind = ln.length - ln.trimStart().length;
    // Walk up at most 40 lines for the nearest opener at a smaller indent
    const minJ = Math.max(-1, i - 40);
    for (let j = i - 1; j > minJ; j--) {
      const p = lines[j];
      const s = p.trim();
      if (!s) continue;
      const pind = p.length - p.trimStart().length;
      if (pind >= ind) continue;
      if (/\)\}\s*$/.test(s) || s.startsWith('</')) {
        break; // left the conditional block
      }
      if (/\{[^}]*(&&|\?)\s*\($/.test(s) || /^\{[^}]*&&\s*</.test(s)) {
        out.push({ tid, line: j + 1, testidLine: i + 1, ctx: s.slice(0, 70) });
        break;
      }
      if (s.endsWith('>') || s.endsWith('(')) {
        ind = pind; // ordinary parent element; keep climbing
      }
    }
  }
  return out;
}

/**
 * Recursively walks a directory matching files against a predicate.
 */
async function walk(dir, matcher) {
  const files = [];
  try {
    const entries = await readdir(dir, { withFileTypes: true });
    for (const e of entries) {
      const full = join(dir, e.name);
      if (e.isDirectory()) {
        files.push(...(await walk(full, matcher)));
      } else if (matcher(e.name, full)) {
        files.push(full);
      }
    }
  } catch {
    // Directory might not exist in fixture
  }
  return files;
}

// Pages are where this bit first, but a conditional block in a shared component is the
// same defect on a second surface -- a guard that watches only one of two surfaces lets
// the next one through. `components/` is nested, so it needs the recursive walk.
const isSource = (name) => name.endsWith('.tsx') && !name.endsWith('.test.tsx');
const sourceFiles = [...(await walk(PAGES_DIR, isSource)), ...(await walk(COMPONENTS_DIR, isSource))].sort();

const testFiles = await walk(SRC_DIR, (name) => /\.test\.tsx?$/.test(name));
const e2eFiles = await walk(E2E_DIR, (name) => /\.ts$/.test(name));
const testAndE2eFiles = [...testFiles, ...e2eFiles];

const refs = new Set();
// リポジトリ内に mrsMop-hint-live や mrsMop-kbd-shortcuts など camelCase を含む testid が存在する。
// 将来これらが条件付きブロックに入った際、テスト側で参照されていても小文字のみの正規表現だと
// 参照として拾えず未参照と誤判定される偽陽性を防ぐため、英大文字も含めて抽出する。
const REF_PATTERN = /["'`]([A-Za-z0-9][A-Za-z0-9_-]{2,})["'`]/g;
for (const f of testAndE2eFiles) {
  const content = await readFile(f, 'utf8');
  for (const m of content.matchAll(REF_PATTERN)) {
    refs.add(m[1]);
  }
}

let conditionalTestidsCount = 0;
const violations = [];

for (const pagePath of sourceFiles) {
  const content = await readFile(pagePath, 'utf8');
  const items = extractConditionalTestids(content);
  for (const item of items) {
    if (item.tid.includes('$') || item.tid.includes('{')) continue;
    conditionalTestidsCount += 1;
    if (!refs.has(item.tid)) {
      violations.push({
        file: relative(FRONTEND, pagePath),
        line: item.line,
        testidLine: item.testidLine,
        tid: item.tid,
      });
    }
  }
}

const filesScanned = sourceFiles.length;
const conditionalTestids = conditionalTestidsCount;

if (SCANNING_REPO) {
  assertFloor('conditional-testid-coverage', filesScanned, 340, 'page and component sources');
  assertFloor('conditional-testid-coverage', conditionalTestids, 800, 'conditional testids');
}

if (violations.length > 0) {
  console.error(`\nconditional-testid-coverage: ${violations.length} unreferenced conditional testid(s) found:\n`);
  for (const v of violations) {
    console.error(`  ${v.file}:${v.line} (testid at line ${v.testidLine}) ${v.tid}`);
  }
  console.error(
    '\nConditional testids must be referenced by at least one test (`src/**/*.test.ts(x)`) or E2E (`e2e/**/*.ts`).\n' +
      'Without a reference, the conditional block may never be rendered or tested.\n',
  );
  process.exit(1);
}

console.log(
  `conditional-testid-coverage: OK (${filesScanned} page/component sources scanned, ${conditionalTestids} conditional testids checked; all referenced).`,
);
