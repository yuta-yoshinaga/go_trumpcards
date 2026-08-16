#!/usr/bin/env bun
// Guard that a page offering hints in the GUI also offers them in the CLI.
//
// Switching a game to CLI mode used to lose the feature silently: the page
// wires `useGameHint` and shows a HintToggle, but the CLI had no `hint`
// command, so typing it answered "Unknown command". 121 games were in that
// state (#5473).
//
// There are two legitimate ways to answer, and a page needs exactly one:
//
//   - the game's parser module handles `hint` and forwards it to the backend
//     (#5791, #5792) -- available when the Web controller dispatches hint;
//   - the page passes `localCommand` using `hintCliText`, answering from the
//     client-side `useGameHint` result (#5793) -- for games whose backend has
//     no hint action at all.
//
// Neither half is visible to the other: a page can wire `localCommand` while
// its parser rejects the word, and a parser can accept `hint` while the page
// never renders one. This checks the property the player actually experiences
// -- that *some* path exists -- rather than either mechanism.

import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

// CLI_HINT_GUARD_ROOT lets the guard's own test point it at a fixture tree.
// Without it the guard could only be tested against the real repo, where every
// case passes and a broken detector would look identical to a clean codebase.
const FRONTEND = process.env.CLI_HINT_GUARD_ROOT ?? fileURLToPath(new URL('..', import.meta.url));
const PAGES = join(FRONTEND, 'src/pages');
const COMMANDS = join(FRONTEND, 'src/utils/cli/commands');
const FIXTURE = process.env.CLI_HINT_GUARD_ROOT !== undefined;

/**
 * Pages exempt from the rule, each with the reason it cannot apply.
 *
 * An entry is a claim that the page has no CLI hint to expose, not a to-do.
 */
const ALLOWED = new Map([
  // Empty on purpose. The first version exempted Memory on the claim that it
  // never receives a hint -- wrong: it computes `frontendHint` from
  // getMemoryHint() instead of taking it from useGameHint, so the exemption was
  // hiding a real gap rather than describing one. An entry here must be a claim
  // that the page has no hint at all, verified against how it renders one.
]);

/** Reads a directory, returning [] when it does not exist. */
async function listFiles(dir) {
  try {
    return await readdir(dir);
  } catch {
    return [];
  }
}

const pageFiles = (await listFiles(PAGES)).filter((f) => f.endsWith('Page.tsx'));
if (!FIXTURE) assertFloor('cli-hint-reachable', pageFiles.length, 200, 'game pages scanned');

const commandFiles = (await listFiles(COMMANDS)).filter((f) => f.endsWith('Commands.ts') && !f.endsWith('.test.ts'));
if (!FIXTURE) assertFloor('cli-hint-reachable', commandFiles.length, 180, 'CLI command modules scanned');

// Modules that answer `hint` themselves, plus the ones they re-export it through.
const answersHint = new Set();
const shared = new Set();
for (const f of commandFiles) {
  const src = await readFile(join(COMMANDS, f), 'utf8');
  if (src.includes("'hint'")) {
    answersHint.add(f.replace(/Commands\.ts$/, ''));
    if (f.startsWith('shared')) shared.add(f.replace(/\.ts$/, ''));
  }
}

let checked = 0;
const missing = [];
for (const f of pageFiles) {
  const src = await readFile(join(PAGES, f), 'utf8');
  const name = f.replace(/Page\.tsx$/, '');

  // Only pages that actually have a hint to show, and a CLI to show it in.
  if (!src.includes('parseCommand:')) continue;

  // "Has a hint" is not the same as "destructures `hint` from useGameHint".
  // MemoryPage computes its own via getMemoryHint() and renders it through
  // FrontendHintTooltip; keying on the destructure skipped it silently, so the
  // guard reported every page reachable while never looking at it.
  const hasHint =
    /const\s*\{[^}]*\bhint\b[^}]*\}\s*=\s*useGameHint/.test(src) ||
    /\bhint:\s*\w+[^}]*\}\s*=\s*useGameHint/.test(src) ||
    /FrontendHintTooltip/.test(src) ||
    /\bget\w+Hint\(/.test(src);
  if (!hasHint) continue;
  if (ALLOWED.has(name)) continue;
  checked += 1;

  // (a) the page answers locally.
  //
  // Both halves are required: an `import { hintCliText }` on its own satisfied
  // the first version of this check, so deleting the localCommand line left the
  // guard green with a dangling import. Verified by removing that line from
  // MemoryPage -- the guard has to notice.
  if (/localCommand:/.test(src) && /hintCliText\(/.test(src)) continue;

  // (b) the page defines its parser inline and handles hint there. Several
  // solitaire pages keep `parseXCommand` in the page file rather than in a
  // commands module, so looking only at src/utils/cli/commands/ reports them
  // as missing when they have handled hint all along.
  if (/function parse\w+Command/.test(src) && /case 'hint':/.test(src)) continue;

  // (c) its parser module answers, directly or via a shared module it imports
  const mod = src.match(/from '\.\.\/utils\/cli\/commands\/(\w+)Commands'/);
  if (mod && answersHint.has(mod[1])) continue;
  if (mod) {
    const modSrc = await readFile(join(COMMANDS, `${mod[1]}Commands.ts`), 'utf8').catch(() => '');
    if ([...shared].some((s) => modSrc.includes(`./${s}`))) continue;
  }

  missing.push(name);
}

if (!FIXTURE) assertFloor('cli-hint-reachable', checked, 150, 'pages with both a hint and a CLI');

if (missing.length > 0) {
  console.error(
    `cli-hint-reachable: ${missing.length} page(s) show hints in the GUI but expose no way to ask ` +
      "for one in CLI mode. Either add a `hint` case to the game's *Commands.ts (when its Web " +
      'controller dispatches hint) or pass `localCommand` with `hintCliText` (when it does not):\n' +
      missing.map((m) => `  - ${m}Page.tsx`).join('\n'),
  );
  process.exit(1);
}

console.log(
  `cli-hint-reachable: OK (${checked} of ${pageFiles.length} pages have both a hint and a CLI; ` +
    `all reachable, ${ALLOWED.size} documented exemption(s)).`,
);
