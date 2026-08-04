#!/usr/bin/env bun
/**
 * Query J-PlatPat (Japan Patent Office) for trademark registrations.
 *
 * Kept in the repository because the 2026-08-03 audit had to write it from
 * scratch, and the check is not a one-off: `21+3` was registered on 2023-11-09
 * and this project added the side bet that displayed it on 2026-03-04, so a
 * mark can already be live years before the code that trips over it lands.
 * Re-run roughly annually, and whenever adding a game or feature named after a
 * commercial product. See TRADEMARKS.md.
 *
 * Two lessons are baked into the interface:
 *
 *   1. **Search the rights holder, not only the mark.** Every mark-name query
 *      for the casino table games returned nothing. `21+3` surfaced only from
 *      `--owner ギャラクシーゲーミング`. Querying what you already suspect
 *      cannot find what you have not thought of, so `--owner` exists and is the
 *      more valuable of the two.
 *   2. **Spaces mean OR, not phrase.** J-PlatPat's own note says so, and
 *      `LET IT RIDE` duly returned every mark containing LET, IT or RIDE.
 *      Pass multi-word marks with the spaces removed: `LETITRIDE`.
 *
 * J-PlatPat is a single-page app with no public API, so this drives a real
 * browser using the Playwright already installed for the E2E suite.
 *
 * Usage:
 *   bun scripts/jpo-trademark-search.mjs --mark LETITRIDE --mark TICHU
 *   bun scripts/jpo-trademark-search.mjs --owner ギャラクシーゲーミング
 *   bun scripts/jpo-trademark-search.mjs --mark ROOK --owner ハズブロ   # AND
 *   bun scripts/jpo-trademark-search.mjs --file names.txt --classes 9,28,41
 *
 * A long sweep must not be trusted on its own: a session that starts returning
 * empty pages looks exactly like a clean result set. `--file` therefore injects
 * a known-positive canary every CANARY_EVERY queries and reports whether each
 * one still returns rows. A run with a failed canary is void, not clean.
 *
 * Sanity-check the harness itself before trusting a run of zeros:
 *   bun scripts/jpo-trademark-search.mjs --mark ポケモン    # expect 任天堂
 *   bun scripts/jpo-trademark-search.mjs --mark COCACOLA   # expect コカ・コーラ
 */

import { readFileSync } from 'node:fs';
import { chromium } from '../frontend/node_modules/playwright-core/index.mjs';

/** Trademark search screen. */
const SEARCH_URL = 'https://www.j-platpat.inpit.go.jp/t0100';
/** 商標(検索用) — the mark-text field. */
const MARK_FIELD = '#t01_srchCondtn_mk_txtKeywd0';
/** 出願人/権利者/名義人 — the rights-holder field. */
const OWNER_FIELD = '#t01_srchCondtn_other_txtKeywd0';
/** The 検索 button, which is an anchor rather than a <button>. */
const SEARCH_BUTTON = 'a#t01_srchBtn_btnSearch';

/** The SPA settles well after `networkidle`; these were measured, not guessed. */
const SETTLE_MS = 3500;
const RESULTS_MS = 10000;

/** A mark whose Japanese registration is certain, used to prove the run is live. */
const CANARY = 'ポケモン';
/** Queries between canaries during a --file sweep. */
const CANARY_EVERY = 40;

function parseArgs(argv) {
  const marks = [];
  const owners = [];
  let classes = null;
  for (let i = 0; i < argv.length; i += 2) {
    const [flag, value] = [argv[i], argv[i + 1]];
    if (!value) throw new Error(`missing value for ${flag}`);
    if (flag === '--mark') marks.push(value);
    else if (flag === '--owner') owners.push(value);
    else if (flag === '--classes') classes = value.split(',').map((c) => c.trim());
    else if (flag === '--file') {
      for (const line of readFileSync(value, 'utf8').split('\n')) {
        const name = line.trim();
        if (name && !name.startsWith('#')) marks.push(name);
      }
    } else throw new Error(`unknown flag ${flag} (expected --mark, --owner, --classes or --file)`);
  }
  return { marks, owners, classes };
}

/**
 * True when a result row is in one of the classes we care about. The class
 * column is the 5th pipe-separated field, space-separated. Filtering matters
 * because most collisions are irrelevant: the only `LET IT RIDE` registration
 * in Japan is class 25 apparel, and `ティチュー` is class 01/03 chemicals.
 */
function inClasses(row, classes) {
  if (!classes) return true;
  const field = (row.split('|')[4] || '').trim();
  // The class column is identified by position, so a layout change would make
  // this filter drop or keep rows for no reason — and these results feed legal
  // judgement calls. Rather than guess, flag the row, show it, and make the run
  // exit non-zero so the operator knows the filter stopped being trustworthy.
  if (!/^\d{1,2}( \d{1,2})*( …)?$/.test(field)) {
    console.log(`  !! class column not recognised (${JSON.stringify(field).slice(0, 40)}) — row shown unfiltered`);
    layoutSuspect = true;
    return true;
  }
  const found = field.split(/\s+/);
  return classes.some((c) => found.includes(c) || found.includes(c.padStart(2, '0')));
}

/** Set when a row's class column did not look like a class list. */
let layoutSuspect = false;

const { marks, owners, classes } = parseArgs(process.argv.slice(2));
if (marks.length === 0 && owners.length === 0) {
  console.error('usage: bun scripts/jpo-trademark-search.mjs [--mark TEXT]... [--owner TEXT]...');
  process.exit(2);
}

// One query per mark and per owner; if both are given once, they are ANDed into
// a single query instead, which is how "does Hasbro hold ROOK in Japan" is asked.
const queries =
  marks.length === 1 && owners.length === 1
    ? [{ mark: marks[0], owner: owners[0] }]
    : [...marks.map((mark) => ({ mark, owner: '' })), ...owners.map((owner) => ({ mark: '', owner }))];

const browser = await chromium.launch({ headless: true });
const page = await (await browser.newContext({ locale: 'ja-JP' })).newPage();

/** Runs one query and returns the result rows of page 1. */
async function search(mark, owner) {
  await page.goto(SEARCH_URL, { waitUntil: 'networkidle', timeout: 90000 });
  await page.waitForTimeout(SETTLE_MS);
  if (mark) await page.fill(MARK_FIELD, mark);
  if (owner) await page.fill(OWNER_FIELD, owner);
  await page.click(SEARCH_BUTTON);
  await page.waitForTimeout(RESULTS_MS);
  return page.evaluate(() =>
    [...document.querySelectorAll('table tr')]
      .map((tr) =>
        [...tr.querySelectorAll('td')]
          .map((td) => (td.innerText || '').replace(/\s+/g, ' ').trim())
          .filter(Boolean)
          .join(' | '),
      )
      .filter((s) => /登録\d|商願/.test(s)),
  );
}

let done = 0;
let canariesRun = 0;
let canariesFailed = 0;

for (const { mark, owner } of queries) {
  const label = `mark="${mark || '-'}" owner="${owner || '-'}"`;
  try {
    const rows = await search(mark, owner);
    const shown = rows.filter((row) => inClasses(row, classes));
    const suffix = classes ? ` (${shown.length} in class ${classes.join('/')} of ${rows.length})` : '';
    console.log(`\n### ${label} — ${rows.length} row(s) on page 1${suffix}`);
    if (shown.length === 0) console.log(classes && rows.length > 0 ? '  該当区分なし' : '  該当なし');
    for (const row of shown.slice(0, 15)) console.log(`  ${row.slice(0, 190)}`);
    if (shown.length > 15) console.log(`  … ${shown.length - 15} more on this page; open J-PlatPat for the rest`);
  } catch (err) {
    // Reported, not thrown: one failed query must not discard the results of
    // the others in a long sweep. A run with errors is not a clean run.
    console.log(`\n### ${label}\n  ERROR ${String(err).split('\n')[0].slice(0, 120)}`);
  }

  done += 1;
  // `>=`, not `>`: a sweep of exactly CANARY_EVERY names reaches the interval
  // and must still be verified. The final query is always canaried too —
  // otherwise the trailing partial chunk (241-264 of a 264-name run) is
  // reported without ever confirming the session was still alive for it.
  const atInterval = queries.length >= CANARY_EVERY && done % CANARY_EVERY === 0;
  const atEnd = queries.length >= CANARY_EVERY && done === queries.length;
  if (atInterval || atEnd) {
    canariesRun += 1;
    let alive = false;
    try {
      alive = (await search(CANARY, '')).length > 0;
    } catch {
      alive = false;
    }
    if (!alive) canariesFailed += 1;
    console.log(`\n--- canary after ${done}: "${CANARY}" ${alive ? 'returned rows (session live)' : 'RETURNED NOTHING — results after this point are not trustworthy'} ---`);
    // A failed canary already voids the run, so the remaining queries would be
    // discarded anyway. Stop rather than spend another 13s each hammering
    // J-PlatPat for results nobody may use.
    if (!alive) {
      console.log(`  stopping early: ${queries.length - done} quer(ies) not run`);
      break;
    }
  }
}

if (layoutSuspect) {
  console.error('\nAt least one class column was unrecognised: the --classes filter cannot be trusted for this run.');
}

if (canariesRun > 0) {
  console.log(`\n=== ${done} quer(ies), ${canariesRun} canary check(s), ${canariesFailed} failed ===`);
  if (canariesFailed > 0) {
    console.error('A canary returned nothing: treat this sweep as void, not clean, and re-run.');
    await browser.close();
    process.exit(1);
  }
}

await browser.close();
if (layoutSuspect) process.exit(1);
