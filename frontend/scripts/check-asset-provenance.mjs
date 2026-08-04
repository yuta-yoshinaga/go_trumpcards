#!/usr/bin/env bun
/**
 * Fails the build when a bundled image has no recorded source.
 *
 * The card set this replaced arrived in the initial commit with no provenance
 * at all and survived six years and hundreds of PRs, because nothing ever
 * asked. Shipping an undocumented asset under this repository's MIT licence
 * grants downstream users rights we may not hold, so the manifest is a gate
 * rather than a convention.
 *
 * `public/images/README.md` owns the manifest; this script only enforces it.
 * The check runs both ways — an undocumented file fails, and so does a
 * manifest entry whose file is gone, because a stale manifest is how a
 * document starts lying.
 */

import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

/** Repository root — this file lives at frontend/scripts/. */
const REPO_ROOT = new URL('../../', import.meta.url).pathname;
const IMAGES_DIR = join(REPO_ROOT, 'public/images');
const MANIFEST_MD = join(IMAGES_DIR, 'README.md');

/**
 * Sanity floor. A walk that finds nothing exits 0 and looks identical to a
 * fully documented tree; the deck alone is 55 files.
 */
const MIN_FILES = 40;

/** Expands one manifest name, which may be a `aNN..aMM.ext` range. */
function expand(name) {
  const range = name.match(/^([a-z]+)(\d+)\.\.[a-z]*(\d+)(\.[a-z0-9]+)$/i);
  if (!range) return [name];
  const [, prefix, from, to, ext] = range;
  const width = from.length;
  const out = [];
  for (let i = Number(from); i <= Number(to); i += 1) {
    out.push(`${prefix}${String(i).padStart(width, '0')}${ext}`);
  }
  return out;
}

const md = readFileSync(MANIFEST_MD, 'utf8');
const block = md.match(/<!-- asset-manifest:start([\s\S]*?)-->/);
if (!block) {
  console.error('check-asset-provenance: no asset-manifest block in public/images/README.md.');
  process.exit(1);
}

const declared = new Set(
  block[1]
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.includes('|'))
    .flatMap((line) => expand(line.split('|')[0].trim())),
);

assertFloor('asset-provenance', declared.size, MIN_FILES, 'files declared in the manifest');

const present = new Set(readdirSync(IMAGES_DIR));

const undocumented = [...present].filter((f) => !declared.has(f)).sort();
const missing = [...declared].filter((f) => !present.has(f)).sort();

if (undocumented.length > 0) {
  console.error(`check-asset-provenance: ${undocumented.length} file(s) with no recorded source:`);
  for (const f of undocumented) console.error(`  - public/images/${f}`);
  console.error(
    '\nAdd each to the asset-manifest block in public/images/README.md with its author and licence.' +
      '\nOnly CC0 or public-domain art may be added — "royalty-free" is not a licence and usually' +
      '\nforbids the redistribution that shipping under MIT performs.',
  );
}

if (missing.length > 0) {
  console.error(`check-asset-provenance: ${missing.length} manifest entry/entries with no file:`);
  for (const f of missing) console.error(`  - ${f}`);
  console.error('\nRemove the stale line from public/images/README.md.');
}

if (undocumented.length > 0 || missing.length > 0) process.exit(1);

console.log(`asset-provenance: OK (${present.size} files in public/images, all with a recorded source).`);
