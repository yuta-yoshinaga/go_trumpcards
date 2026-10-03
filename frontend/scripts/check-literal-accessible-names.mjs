#!/usr/bin/env bun
/** Reject English string literals used as accessible-name attributes in JSX. */
import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const ROOT = resolve(process.cwd());
const SOURCE_DIRS = ['src/pages', 'src/components'].map((path) => join(ROOT, path));
const JSX_EXTENSIONS = new Set(['.tsx']);
const ATTRIBUTE = /(?:aria-label|title|alt)\s*=\s*(["'])([A-Za-z][^"']*)\1/g;

/** Return English string literal attributes found in JSX source. */
export function findLiteralAccessibleNames(source) {
  return [...source.matchAll(ATTRIBUTE)].map(([match]) => match);
}

function collectTsxFiles(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return collectTsxFiles(path);
    return JSX_EXTENSIONS.has(path.slice(path.lastIndexOf('.'))) && !/\.test\.tsx$/.test(path) ? [path] : [];
  });
}

if (import.meta.main) {
  const files = SOURCE_DIRS.flatMap(collectTsxFiles);
  assertFloor('literal-accessible-names', files.length, 300, 'TSX source files scanned');
  const violations = files.flatMap((file) =>
    findLiteralAccessibleNames(readFileSync(file, 'utf8')).map((attribute) => `${file}: ${attribute}`),
  );

  if (violations.length) {
    console.error(`literal-accessible-names: ${violations.length} English literal(s) found.`);
    for (const violation of violations) console.error(`  ${violation}`);
    process.exit(1);
  }
  console.log(`literal-accessible-names: OK (${files.length} TSX source files scanned; 0 English literals).`);
}
