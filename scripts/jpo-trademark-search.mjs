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
 *
 * Sanity-check the harness itself before trusting a run of zeros:
 *   bun scripts/jpo-trademark-search.mjs --mark ポケモン    # expect 任天堂
 *   bun scripts/jpo-trademark-search.mjs --mark COCACOLA   # expect コカ・コーラ
 */

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

function parseArgs(argv) {
  const marks = [];
  const owners = [];
  for (let i = 0; i < argv.length; i += 2) {
    const [flag, value] = [argv[i], argv[i + 1]];
    if (!value) throw new Error(`missing value for ${flag}`);
    if (flag === '--mark') marks.push(value);
    else if (flag === '--owner') owners.push(value);
    else throw new Error(`unknown flag ${flag} (expected --mark or --owner)`);
  }
  return { marks, owners };
}

const { marks, owners } = parseArgs(process.argv.slice(2));
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

for (const { mark, owner } of queries) {
  const label = `mark="${mark || '-'}" owner="${owner || '-'}"`;
  try {
    await page.goto(SEARCH_URL, { waitUntil: 'networkidle', timeout: 90000 });
    await page.waitForTimeout(SETTLE_MS);
    if (mark) await page.fill(MARK_FIELD, mark);
    if (owner) await page.fill(OWNER_FIELD, owner);
    await page.click(SEARCH_BUTTON);
    await page.waitForTimeout(RESULTS_MS);

    const rows = await page.evaluate(() =>
      [...document.querySelectorAll('table tr')]
        .map((tr) =>
          [...tr.querySelectorAll('td')]
            .map((td) => (td.innerText || '').replace(/\s+/g, ' ').trim())
            .filter(Boolean)
            .join(' | '),
        )
        .filter((s) => /登録\d|商願/.test(s)),
    );

    console.log(`\n### ${label} — ${rows.length} row(s) on page 1`);
    if (rows.length === 0) console.log('  該当なし');
    for (const row of rows.slice(0, 15)) console.log(`  ${row.slice(0, 190)}`);
    if (rows.length > 15) console.log(`  … ${rows.length - 15} more on this page; open J-PlatPat for the rest`);
  } catch (err) {
    // Reported, not thrown: one failed query must not discard the results of
    // the others in a long sweep. A run with errors is not a clean run.
    console.log(`\n### ${label}\n  ERROR ${String(err).split('\n')[0].slice(0, 120)}`);
  }
}

await browser.close();
