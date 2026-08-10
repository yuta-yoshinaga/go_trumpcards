#!/usr/bin/env bun
// Guard that every ```mermaid block in the repo's Markdown actually parses.
//
// Usage: check-mermaid.mjs [dir]     (default: the repository root)
//
// `docs/design/{backend,frontend}.md` are almost entirely Mermaid — 140 blocks
// between them — and nothing has ever validated them. A diagram that fails to
// parse renders on GitHub as a red error box, and because no tool reads `.md`
// (tsc, Biome and vitest all skip it) the break survives review and ships.
//
// The 2026-08-08 drift sweep (#5170) rewrote roughly 350 lines across those two
// files: deleted classes, renamed methods, restructured relations. That kind of
// bulk edit is exactly where a stray `}` or a broken arrow gets introduced, and
// exactly where "it looked fine in the diff" is not evidence. This check runs
// the real `mermaid.parse()` so a malformed diagram fails the build instead.
// On its first run it found four already-broken diagrams that were rendering as
// error boxes in the published manuals.
//
// Parsing needs a DOM. `mermaid` is a runtime dependency and `jsdom` a dev
// dependency, so both are declared — this does not reach for anything ambient.
// Note `globalThis.navigator` has no setter under Node 24 and must be left
// alone; assigning to it throws before mermaid is ever imported.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { JSDOM } from 'jsdom';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');

/** Scan target: a caller-supplied directory (tests) or the whole repo. */
const ROOT = process.argv[2] ? resolve(process.argv[2]) : REPO;
const SCANNING_REPO = !process.argv[2];

/** Directories that are not ours to police. */
const SKIP = new Set(['node_modules', '.git', 'dist', 'coverage', 'playwright-report', 'test-results']);

/** Every ```mermaid fenced block, captured without its fences. */
const MERMAID_BLOCK = /^```mermaid\r?\n([\s\S]*?)^```/gm;

// The files this guard primarily exists to protect. A repo-wide floor cannot
// defend them: they hold 140 of ~646 blocks, so losing both of them entirely
// still leaves the repo-wide totals far above any sane floor, and the guard
// would report green while the worst drift site went unchecked. That is the
// same silent-zero failure this script was written to prevent, so the two files
// get a floor of their own.
const DESIGN_DOCS = ['docs/design/backend.md', 'docs/design/frontend.md'];

async function* markdownFiles(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (SKIP.has(entry.name)) continue;
    const full = join(dir, entry.name);
    if (entry.isDirectory()) yield* markdownFiles(full);
    else if (entry.name.endsWith('.md')) yield full;
  }
}

const dom = new JSDOM('<!doctype html><html><body></body></html>');
globalThis.window = dom.window;
globalThis.document = dom.window.document;
globalThis.HTMLElement = dom.window.HTMLElement;

const { default: mermaid } = await import('mermaid');
mermaid.initialize({ startOnLoad: false });

let blocks = 0;
let files = 0;
let designBlocks = 0;
const failures = [];

for await (const file of markdownFiles(ROOT)) {
  const src = await readFile(file, 'utf8');
  const found = [...src.matchAll(MERMAID_BLOCK)];
  if (found.length === 0) continue;
  files++;
  const rel = relative(ROOT, file).split('\\').join('/');
  if (DESIGN_DOCS.includes(rel)) designBlocks += found.length;
  for (const [index, match] of found.entries()) {
    blocks++;
    try {
      await mermaid.parse(match[1]);
    } catch (error) {
      const line = src.slice(0, match.index).split('\n').length;
      const reason = String(error?.message ?? error).split('\n')[0];
      failures.push(`${rel}:${line} (block #${index + 1}): ${reason}`);
    }
  }
}

// Floors apply to the real corpus only; a caller-supplied fixture directory is
// deliberately small. Sized at roughly two thirds of the current counts, per
// the convention in scripts/lib/floor.mjs.
if (SCANNING_REPO) {
  assertFloor('mermaid', files, 340, 'files containing mermaid blocks');
  assertFloor('mermaid', blocks, 430, 'mermaid blocks');
  assertFloor('mermaid', designBlocks, 90, 'mermaid blocks in docs/design');
}

if (failures.length > 0) {
  console.error(`\nmermaid: ${failures.length} block(s) failed to parse.\n`);
  for (const failure of failures) console.error(`  ${failure}`);
  console.error('\nThese render as an error box on GitHub. Fix the diagram syntax.\n');
  process.exit(1);
}

console.log(`mermaid: OK (${blocks} blocks across ${files} files, ${designBlocks} in docs/design).`);
