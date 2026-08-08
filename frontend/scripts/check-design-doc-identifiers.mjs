#!/usr/bin/env bun
// Guard that every class named in a docs/design/frontend.md class diagram still
// exists in frontend/src.
//
// Usage: check-design-doc-identifiers.mjs [srcDir] [docFile]
//
// The Go half of this pair lives in
// internal/infrastructure/games/design_doc_identifiers_test.go. Together they
// close the last hole found by the 2026-08-08 drift sweep (#5170): the two
// design documents were the worst drift site in the repo and nothing checked
// their identifiers. On the frontend side the sweep found four hooks that no
// longer exist at all (`useGameSound`, `useCardGesture`, `useHaptics`,
// `useBlackJackGame` and friends) still drawn as classes.
//
// Scope: class names only, not members. A hook's return shape and an object
// literal's keys cannot be resolved without full type information, and a guard
// that reports false positives gets switched off. Renamed *members* remain
// unguarded here; renamed or deleted *components, hooks and modules* — which is
// what actually rotted — do not.
//
// A diagram node resolves if it is any of:
//   1. an exported binding somewhere under src/,
//   2. a module: src/**/<name>.ts(x), or a directory src/**/<name>/,
//   3. a listed placeholder (a node standing in for "the per-game page").

import { readdir, readFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');

const SRC = process.argv[2] ? resolve(process.argv[2]) : join(FRONTEND, 'src');
const DOC = process.argv[3] ? resolve(process.argv[3]) : join(REPO, 'docs/design/frontend.md');
const SCANNING_REPO = !process.argv[2];

/** Nodes that deliberately stand for a family rather than one symbol. */
const PLACEHOLDERS = new Set([
  // One node standing for all 264 game pages.
  'GamePage',
  // Page-local phase constants, not exported enums. The document says so
  // itself where it lists them.
  'DurakPhase',
  'PigsTailPhase',
]);

const SKIP = new Set(['node_modules', '.git', 'dist', 'coverage', 'playwright-report', 'test-results']);

/** `export function X` / `export const X` / `export class X` / `export type X` … */
const EXPORT_DECL =
  /export\s+(?:default\s+)?(?:async\s+)?(?:function|const|let|var|class|interface|type|enum)\s+([A-Za-z0-9_$]+)/g;
/**
 * `export { A, B as C }` — take the exported (right-hand) name. The optional
 * `type` keyword matters: `export type { A as B }` is a real re-export, and
 * missing it would report a legitimately exported name as unresolved.
 */
const EXPORT_LIST = /export\s*(?:type\s+)?\{([^}]*)\}/g;
/** `export default someIdentifier` — a bare re-export of an existing binding. */
const EXPORT_DEFAULT_IDENT = /export\s+default\s+([A-Za-z0-9_$]+)\s*;/g;

const CLASS_DECL = /^\s*class\s+([A-Za-z0-9_]+)/;

async function* sourceFiles(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (SKIP.has(entry.name)) continue;
    const full = join(dir, entry.name);
    if (entry.isDirectory()) yield* sourceFiles(full);
    else if (/\.(ts|tsx)$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) yield full;
  }
}

async function* directories(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (SKIP.has(entry.name) || !entry.isDirectory()) continue;
    const full = join(dir, entry.name);
    yield entry.name;
    yield* directories(full);
  }
}

const known = new Set();
let sourceCount = 0;

for await (const file of sourceFiles(SRC)) {
  sourceCount++;
  const src = await readFile(file, 'utf8');
  // The module itself is addressable as a diagram node.
  known.add(
    file
      .split('/')
      .pop()
      .replace(/\.tsx?$/, ''),
  );
  for (const m of src.matchAll(EXPORT_DECL)) known.add(m[1]);
  for (const m of src.matchAll(EXPORT_DEFAULT_IDENT)) known.add(m[1]);
  for (const m of src.matchAll(EXPORT_LIST)) {
    for (const part of m[1].split(',')) {
      const name = part
        .trim()
        .split(/\s+as\s+/)
        .pop()
        ?.trim();
      if (name) known.add(name);
    }
  }
}
for await (const name of directories(SRC)) known.add(name);

const doc = await readFile(DOC, 'utf8');
const classes = new Set();
for (const block of doc.matchAll(/```mermaid\n([\s\S]*?)```/g)) {
  if (!/^\s*classDiagram/m.test(block[1])) continue;
  for (const line of block[1].split('\n')) {
    const m = line.match(CLASS_DECL);
    if (m) classes.add(m[1]);
  }
}

if (SCANNING_REPO) {
  assertFloor('design-doc-identifiers', sourceCount, 700, 'source files scanned');
  assertFloor('design-doc-identifiers', known.size, 3000, 'known identifiers');
  assertFloor('design-doc-identifiers', classes.size, 140, 'diagram classes');
}

const unknown = [...classes].filter((c) => !PLACEHOLDERS.has(c) && !known.has(c)).sort();

if (unknown.length > 0) {
  console.error(
    `\ndesign-doc-identifiers: ${unknown.length} class(es) in the design doc have no counterpart in src/.\n`,
  );
  for (const name of unknown) console.error(`  ${name}`);
  console.error(
    '\nEither the symbol was renamed or deleted (update the diagram), or it is a\n' +
      'deliberate placeholder (add it to PLACEHOLDERS in this script).\n',
  );
  process.exit(1);
}

console.log(`design-doc-identifiers: OK (${classes.size} diagram classes resolved against ${known.size} identifiers).`);
