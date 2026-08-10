#!/usr/bin/env bun
// Guard that every guard reports a count it has actually floored.
//
// The other scripts in this directory all end the same way: a line saying how much was
// inspected, followed by exit 0. That line is what a reader takes as coverage -- "2020 source
// files scanned", "774 codes", "6212 files". None of it is worth anything if the walk that
// produced the number broke, because a broken walk finds no violations and exits 0 exactly
// like a clean tree. check-dependency-licenses inspected 341 packages and reported 1, and the
// run looked perfect.
//
// Three guards grew `MIN_*` constants after that. This one makes the practice enforceable:
//
//   Any guard that prints a discovered count must assert a floor under it.
//
// "Prints a discovered count" is read mechanically as a `console.log` containing a `${}`
// interpolation -- exactly the lines that make a claim about scope. "Asserts a floor" is a
// call to `assertFloor` from ./lib/floor.mjs. A guard with three such lines needs three
// floors; one floor does not vouch for the other two walks.
//
// Floors of 0 are rejected. `assertFloor(x, 0)` passes for every possible input, which is the
// shape of a guard that was added to satisfy a checker rather than to catch anything.
//
// Usage: check-guard-floors.mjs [dir]     (default: this scripts/ directory)

import { readdir, readFile } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const SCRIPTS = fileURLToPath(new URL('.', import.meta.url));
const REPO = join(SCRIPTS, '../..');
const DIR = process.argv[2] ? process.argv[2] : SCRIPTS;

// This file is excluded from its own scan. It would otherwise count the `assertFloor(` and
// `console.log` fragments that appear here as regex source and prose, and report on itself
// using numbers that describe its own documentation. Its own walk is floored below instead.
const SELF = 'check-guard-floors.mjs';

/** Success lines that state a discovered count: `console.log` with an interpolation. */
function countedLogs(src) {
  // console.log( ... ) spanning lines, up to the matching close at column 0 or `);`
  const logs = [...src.matchAll(/console\.log\(([\s\S]*?)\);/g)].map((m) => m[1]);
  return logs.filter((body) => body.includes('${'));
}

/** Every `assertFloor(label, count, min, ...)` call, with the raw `min` argument. */
function floorCalls(src) {
  const calls = [];
  for (const m of src.matchAll(/assertFloor\(([\s\S]*?)\);/g)) {
    // Split only at top-level commas so a template literal holding a comma stays one argument.
    // Backticks toggle rather than nest — counting them like brackets leaves the depth stuck
    // above zero for the rest of the call, and every later argument merges into one.
    const args = [];
    let depth = 0;
    let inTick = false;
    let current = '';
    for (const ch of m[1]) {
      if (ch === '`') inTick = !inTick;
      else if (!inTick && '([{'.includes(ch)) depth += 1;
      else if (!inTick && ')]}'.includes(ch)) depth -= 1;
      if (ch === ',' && depth === 0 && !inTick) {
        args.push(current.trim());
        current = '';
        continue;
      }
      current += ch;
    }
    args.push(current.trim());
    calls.push({ min: args[2] ?? '', raw: m[0] });
  }
  return calls;
}

/**
 * Resolve a `min` argument to a number.
 *
 * Accepts a literal (`200`) or a same-file `const NAME = 200`. Anything computed at runtime
 * is unresolvable here and is reported rather than waved through -- a floor nobody can read
 * off the source is a floor nobody will notice going to zero.
 */
function resolveMin(arg, src) {
  if (/^\d+$/.test(arg)) return Number(arg);
  if (/^[A-Za-z_$][\w$]*$/.test(arg)) {
    const decl = new RegExp(`const\\s+${arg}\\s*=\\s*(\\d+)\\s*;`).exec(src);
    if (decl) return Number(decl[1]);
  }
  return null;
}

// `check-*.test.mjs` is excluded along with SELF: a guard's test file quotes broken guards as
// fixtures, so scanning it reports the fixtures as real violations.
const files = (await readdir(DIR))
  .filter((f) => f.startsWith('check-') && f.endsWith('.mjs') && !f.endsWith('.test.mjs') && f !== SELF)
  .sort();

const problems = [];
let floors = 0;

for (const file of files) {
  const src = await readFile(join(DIR, file), 'utf8');
  const logs = countedLogs(src);
  const calls = floorCalls(src);
  floors += calls.length;
  const where = relative(REPO, join(DIR, file));

  if (calls.length < logs.length) {
    problems.push(
      `${where}: reports ${logs.length} discovered count(s) but asserts ${calls.length} floor(s).\n` +
        `      Each of these lines claims a scope that nothing verifies:\n` +
        logs
          .slice(calls.length)
          .map((l) => `        ${l.replace(/\s+/g, ' ').trim().slice(0, 96)}`)
          .join('\n'),
    );
  }

  for (const call of calls) {
    const min = resolveMin(call.min, src);
    if (min === null) {
      problems.push(
        `${where}: floor \`${call.min}\` is not a literal or a same-file const.\n` +
          '      Use a number, so the floor can be read (and reviewed) off the source.',
      );
    } else if (min <= 0) {
      problems.push(
        `${where}: floor is ${min}. A floor of 0 holds for every input, including a walk\n` +
          '      that found nothing — which is the case this whole mechanism exists to catch.',
      );
    }
  }
}

// This guard has the failure mode it polices: point it at the wrong directory and it finds no
// scripts, reports no problems, and exits 0. Ten guards live here.
assertFloor('guard-floors', files.length, 8, `guard scripts in ${relative(REPO, DIR)}`);

if (problems.length > 0) {
  console.error('\nGuards that report a count without a floor under it:\n');
  for (const p of problems) console.error(`  ${p}\n`);
  console.error(
    'Import assertFloor from ./lib/floor.mjs and assert the count before reporting it.\n' +
      'Set the floor well below the real number — enough that ordinary churn never trips it,\n' +
      'and a walk that found a fraction of the truth does.\n',
  );
  process.exit(1);
}

console.log(`guard-floors: OK (${files.length} guard scripts, ${floors} floors; every reported count is asserted).`);
