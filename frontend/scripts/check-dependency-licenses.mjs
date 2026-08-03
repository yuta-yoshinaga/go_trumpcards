#!/usr/bin/env bun
/**
 * Fails the build when an installed npm dependency carries a license this
 * project cannot redistribute under MIT.
 *
 * Why hand-rolled instead of `license-checker`: that tool reports
 * `UNLICENSED: 1` against this repo's bun-installed `node_modules` — it walks
 * away having inspected nothing and exits 0, which reads exactly like a clean
 * result. A scanner that silently scans nothing is worse than no scanner, so
 * this one asserts it actually saw a plausible number of packages (see
 * MIN_PACKAGES) and fails loudly if it did not.
 *
 * Go dependencies are covered separately by the `licenses-backend` CI job
 * (`go-licenses check`), which understands build tags and so also covers the
 * js/wasm-only worker dependencies.
 */

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';

/** Root of the installed dependency tree. */
const NODE_MODULES = new URL('../node_modules', import.meta.url).pathname;

/**
 * Sanity floor for the number of packages inspected. The tree currently holds
 * ~340 top-level entries; anything under this means the walk broke rather than
 * that the tree is clean.
 */
const MIN_PACKAGES = 100;

/**
 * Licenses that cannot be sublicensed under MIT. LGPL is included because the
 * frontend is bundled: Vite inlines dependency code into the shipped chunks,
 * which is static linking, not the dynamic linking LGPL carves out.
 */
const DENIED = [
  'GPL-1.0',
  'GPL-2.0',
  'GPL-3.0',
  'AGPL-1.0',
  'AGPL-3.0',
  'LGPL-2.0',
  'LGPL-2.1',
  'LGPL-3.0',
  'SSPL-1.0',
  'CC-BY-SA',
  'CC-BY-NC',
  'EUPL',
  'OSL-3.0',
  'CPAL-1.0',
  'RPL-1.5',
];

/** Licenses that are fine to redistribute under MIT. */
const ALLOWED = [
  'MIT',
  'ISC',
  'Apache-2.0',
  'BSD-2-Clause',
  'BSD-3-Clause',
  '0BSD',
  'CC0-1.0',
  'Unlicense',
  'Python-2.0',
  'BlueOak-1.0.0',
  'MPL-2.0',
  'Artistic-2.0',
  'WTFPL',
  'Zlib',
];

/** Reads a package.json, returning null when it is absent or unparsable. */
function readManifest(dir) {
  try {
    return JSON.parse(readFileSync(join(dir, 'package.json'), 'utf8'));
  } catch {
    return null;
  }
}

/** Normalizes the several shapes npm allows for the license field. */
function licenseOf(manifest) {
  if (typeof manifest.license === 'string') return manifest.license;
  if (manifest.license?.type) return manifest.license.type;
  if (Array.isArray(manifest.licenses)) {
    return manifest.licenses.map((l) => l.type ?? l).join(' OR ');
  }
  return '';
}

/**
 * True when the SPDX expression leaves us no permissive option. A dual license
 * such as `(GPL-2.0 OR MIT)` is fine — we take the MIT branch — so a denied id
 * only counts when no allowed id appears alongside it.
 */
function isBlocked(expr) {
  const upper = expr.toUpperCase();
  const hasDenied = DENIED.some((id) => upper.includes(id.toUpperCase()));
  if (!hasDenied) return false;
  return !ALLOWED.some((id) => upper.includes(id.toUpperCase()));
}

/** Collects every installed package directory, descending into `@scope` dirs. */
function packageDirs(root) {
  const dirs = [];
  for (const entry of readdirSync(root)) {
    if (entry === '.bin' || entry === '.cache') continue;
    const full = join(root, entry);
    if (!statSync(full).isDirectory()) continue;
    if (entry.startsWith('@')) {
      for (const scoped of readdirSync(full)) dirs.push(join(full, scoped));
    } else {
      dirs.push(full);
    }
  }
  return dirs;
}

const blocked = [];
const unknown = [];
let scanned = 0;

for (const dir of packageDirs(NODE_MODULES)) {
  const manifest = readManifest(dir);
  if (!manifest?.name) continue;
  scanned += 1;
  const license = licenseOf(manifest);
  if (!license) {
    unknown.push(manifest.name);
  } else if (isBlocked(license)) {
    blocked.push(`${manifest.name}@${manifest.version ?? '?'}: ${license}`);
  }
}

if (scanned < MIN_PACKAGES) {
  console.error(
    `check-dependency-licenses: only ${scanned} packages inspected (expected >= ${MIN_PACKAGES}).\n` +
      'The scan is broken, not the tree clean. Run `bun install` and re-run.',
  );
  process.exit(1);
}

if (blocked.length > 0) {
  console.error(
    `check-dependency-licenses: ${blocked.length} dependency/dependencies cannot be redistributed under MIT:`,
  );
  for (const line of blocked) console.error(`  - ${line}`);
  console.error('\nReplace the dependency, or document a written grant before shipping it.');
  process.exit(1);
}

if (unknown.length > 0) {
  console.warn(`check-dependency-licenses: ${unknown.length} package(s) declare no license: ${unknown.join(', ')}`);
}

console.log(`check-dependency-licenses: ${scanned} packages inspected, no incompatible licenses.`);
