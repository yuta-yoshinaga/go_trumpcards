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

import { existsSync, readdirSync, readFileSync, realpathSync, statSync } from 'node:fs';
import { join } from 'node:path';

/** Root of the installed dependency tree. */
const NODE_MODULES = new URL('../node_modules', import.meta.url).pathname;

/**
 * Sanity floor for the number of packages inspected. The recursive walk
 * currently sees ~477; anything under this means the walk broke rather than
 * that the tree is clean. Set well below the real count so a dependency
 * removal does not trip it, but far enough above zero to stay diagnostic.
 */
const MIN_PACKAGES = 300;

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

/**
 * Packages whose manifest omits `license` but whose real license was read from
 * the files they ship. Each entry records that evidence, so the exemption can
 * be re-checked rather than trusted.
 */
const UNLICENSED_ALLOWLIST = {
  // node_modules/khroma/license: "The MIT License (MIT) — Copyright (c)
  // 2019-present Fabio Spampinato, Andrew Maney". Upstream simply omits the
  // package.json field. Pulled in by mermaid.
  khroma: 'MIT, per the bundled license file',
};

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

/**
 * Collects every installed package directory: top level, `@scope/*`, and any
 * nested `node_modules` a version conflict produced. The nested case is not
 * hypothetical here — this tree carries 24 such packages, and a top-level-only
 * walk skipped them while still clearing the floor, so the floor would have
 * papered over the gap rather than exposing it.
 */
function packageDirs(root, seen = new Set()) {
  // bun leaves a `node_modules/node_modules` symlink pointing at its own
  // parent, so a naive recursion walks it forever. Skip the name outright and
  // keep a realpath set, because a package may also be symlinked into more
  // than one place.
  const key = realpathSync(root);
  if (seen.has(key)) return [];
  seen.add(key);

  const dirs = [];
  for (const entry of readdirSync(root)) {
    if (entry === '.bin' || entry === '.cache' || entry === 'node_modules') continue;
    const full = join(root, entry);
    if (!statSync(full).isDirectory()) continue;
    // A scoped directory is not itself a package; its children are. Either way
    // each package can carry its own nested tree, so recurse from every one of
    // them — `@testing-library/dom/node_modules` holds two packages that a
    // top-level-only pass never sees.
    const packages = entry.startsWith('@') ? readdirSync(full).map((scoped) => join(full, scoped)) : [full];
    for (const pkg of packages) {
      if (!statSync(pkg).isDirectory()) continue;
      dirs.push(pkg);
      const nested = join(pkg, 'node_modules');
      if (existsSync(nested)) dirs.push(...packageDirs(nested, seen));
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
    if (!(manifest.name in UNLICENSED_ALLOWLIST)) unknown.push(manifest.name);
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
  console.error(`check-dependency-licenses: ${unknown.length} package(s) declare no license:`);
  for (const name of unknown) console.error(`  - ${name}`);
  console.error(
    '\nA package with no license metadata is "all rights reserved" by default, which is the' +
      '\nsame undocumented-provenance risk this repository just spent an audit removing from its' +
      '\nart. Establish the real license (the package usually ships a LICENSE file even when the' +
      '\nmanifest omits the field) and add it to UNLICENSED_ALLOWLIST with that evidence.',
  );
  process.exit(1);
}

console.log(`check-dependency-licenses: ${scanned} packages inspected, no incompatible licenses.`);
