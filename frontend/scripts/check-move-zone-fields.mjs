#!/usr/bin/env bun
// Guard that a game's move-zone field names agree between Go and TypeScript.
// See issue #5289.
//
// The index field is spelled two ways across the repo — `col` in 44 Go
// controllers, `idx` in 7 — and both are correct where they are used. The
// hazard is a *new* game assembled from two templates: take the domain from
// Congress (`col`) and the page from Four Seasons (`idx`) and the two halves
// disagree while every existing check still passes.
//
// Nothing else covers it. The page test mocks the API module and asserts the
// argument object, so it never crosses the wire. The Go controller test builds
// the request body itself, so it agrees with the bug. And tsc does not read Go
// struct tags. #5288 (Colorado) shipped that way and every move with a
// destination came back 400 — only E2E caught it, and only for the one game
// that had a spec exercising a destination.

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
// Roots are overridable by argv so this guard's own test can point it at a
// fixture; a guard that can only run against the real tree cannot be shown to
// fail on bad input.
const GO_DIR = process.argv[2] ?? join(REPO, 'internal/adapter/controller');
const TS_DIRS = process.argv[3]
  ? [process.argv[3]]
  : [join(FRONTEND, 'src/api/games'), join(FRONTEND, 'src/types/games')];
const REAL_FLOOR = 30;
const FIXTURE_FLOOR = 1;

/** Fields of every `type <Game>WebZone struct` in the Go controllers. */
async function goZones(dir) {
  const out = new Map();
  for (const name of (await readdir(dir)).filter((f) => f.endsWith('.go') && !f.endsWith('_test.go'))) {
    const text = await readFile(join(dir, name), 'utf8');
    for (const m of text.matchAll(/type\s+([A-Za-z0-9]+)WebZone\s+struct\s*\{([\s\S]*?)\n\}/g)) {
      const fields = new Set();
      for (const f of m[2].matchAll(/`json:"([^",]+)/g)) {
        if (f[1] !== '-') fields.add(f[1]);
      }
      out.set(m[1], { fields, file: name });
    }
  }
  return out;
}

/** Properties of every `export interface <Game>MoveZone` on the TypeScript side. */
async function tsZones(dirs) {
  const out = new Map();
  for (const dir of dirs) {
    let names;
    try {
      names = (await readdir(dir)).filter((f) => f.endsWith('.ts') && !f.endsWith('.test.ts'));
    } catch {
      continue; // an absent directory is not a finding; the floor below covers a bad root
    }
    for (const name of names) {
      const text = await readFile(join(dir, name), 'utf8');
      for (const m of text.matchAll(/export\s+interface\s+([A-Za-z0-9]+)MoveZone\s*\{([\s\S]*?)\n\}/g)) {
        const fields = new Set();
        for (const f of m[2].matchAll(/^\s*([a-zA-Z0-9_]+)\??\s*:/gm)) fields.add(f[1]);
        out.set(m[1], { fields, file: `${relative(FRONTEND, dir)}/${name}` });
      }
    }
  }
  return out;
}

const go = await goZones(GO_DIR);
const ts = await tsZones(TS_DIRS);
// Both floors are plain literals so the number can be read (and reviewed) off the
// source — check-guard-floors rejects a computed one. A fixture run has a handful of
// structs by design; the real run keeps the number that matters.
if (process.argv[2]) {
  assertFloor('move-zone-fields', go.size, FIXTURE_FLOOR, 'WebZone structs in the fixture');
} else {
  assertFloor('move-zone-fields', go.size, REAL_FLOOR, `<Game>WebZone structs in ${relative(REPO, GO_DIR)}`);
}

const problems = [];
for (const [game, g] of go) {
  const t = ts.get(game);
  // A Go zone with no TypeScript counterpart is not a mismatch: some games
  // never send a zone from the client. Only a pair that exists on both sides
  // can disagree, and that is the case this guard is about.
  if (!t) continue;
  const onlyGo = [...g.fields].filter((f) => !t.fields.has(f));
  const onlyTs = [...t.fields].filter((f) => !g.fields.has(f));
  if (onlyGo.length === 0 && onlyTs.length === 0) continue;
  problems.push(
    `  ${game}: ${[...g.fields].join(', ')}  (Go: ${g.file})\n` +
      `  ${' '.repeat(game.length)}  ${[...t.fields].join(', ')}  (TS: ${t.file})\n` +
      (onlyGo.length > 0 ? `      only in Go: ${onlyGo.join(', ')}\n` : '') +
      (onlyTs.length > 0 ? `      only in TS: ${onlyTs.join(', ')}\n` : ''),
  );
}

if (problems.length > 0) {
  console.error('\nMove-zone fields disagree between Go and TypeScript:\n');
  for (const p of problems) console.error(p);
  console.error(
    `${problems.length} mismatched zone(s). The client's index field must be spelled exactly as the` +
      ' Go struct tag, or the destination never reaches the server and every move with a target 400s.\n' +
      '  See issue #5289.',
  );
  process.exit(1);
}

const paired = [...go.keys()].filter((g) => ts.has(g)).length;
console.log(`move-zone-fields: OK (${paired} of ${go.size} zone struct(s) paired with TypeScript, all fields agree).`);
