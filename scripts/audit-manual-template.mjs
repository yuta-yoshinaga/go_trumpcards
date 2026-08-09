#!/usr/bin/env bun
/**
 * Audits every per-game manual under docs/manual/{cui,web} against the
 * structure that docs/manual/{cui,web}_template.md prescribes.
 *
 * Why it exists: the template contract was prose only. `TestPerGameManualsMatchRegistry`
 * checks that a manual *exists* for each registered game and nothing about what
 * is inside it, so 380 of the 528 manuals had drifted when this script first
 * ran — 110 CUI command tables never mentioned `help`, 106 Web manuals omitted
 * `go run ./cmd/server`, and 12 CUI manuals still told the reader to run
 * `go run ./cmd/cli <game>`, a binary this repo does not have. The manuals are
 * rendered in the Web GUI (frontend/src/constants/{cui,}manualTexts.ts), so
 * those gaps are user-facing.
 *
 * Report totals move as the rules below are tightened or relaxed; the counts
 * above are what this script reported on its first run, and the classes are
 * what matters. Trust a fresh run over any number written in prose.
 *
 * This script is the human-facing worklist. The commit gate is the Go test
 * `TestPerGameManualsFollowTemplate` in
 * internal/infrastructure/games/manual_template_test.go, which asserts the
 * same rules — keep the two in sync.
 *
 * Usage:
 *   bun scripts/audit-manual-template.mjs           # grouped report
 *   bun scripts/audit-manual-template.mjs --json    # machine-readable worklist
 */

import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

/** Repository root — this file lives at scripts/. */
const REPO_ROOT = new URL('../', import.meta.url).pathname;
const MANUAL_DIR = join(REPO_ROOT, 'docs/manual');
const NAV_JA = join(REPO_ROOT, 'frontend/src/i18n/locales/ja/common.json');

/**
 * Sanity floor. A walk that finds nothing reports "everything conforms",
 * which is indistinguishable from success — the failure mode that let
 * check-dependency-licenses report 341 packages as 1.
 */
const MIN_MANUALS = 250;

/**
 * Section contract per manual kind, in the order the template lays them out.
 *
 * 遊び方のコツ is deliberately absent: both templates mark it
 * `<!-- 任意セクション -->`. docs/new-game-checklist.md lists it alongside the
 * required ones, which is the doc to correct -- requiring it here would buy
 * conformance by adding 41 sections of filler strategy prose, which is worse
 * for a reader than an honestly absent section.
 */
const SPEC = {
  cui: {
    suffix: '（CUI版）遊び方',
    sections: ['ゲーム概要', '起動方法', 'ルール', 'ゲームの流れ', 'コマンド一覧', '画面の見方'],
  },
  web: {
    suffix: '（Web版）遊び方',
    sections: ['ゲーム概要', '起動方法', 'ルール', 'ゲームの流れ', '画面の操作方法', '画面構成'],
  },
};

/**
 * Returns the document's lines with everything inside a fenced code block
 * blanked out.
 *
 * Heading detection MUST go through this. A `# 英語表示:` comment inside a
 * ```sh block is not a heading, and counting it as one produced 13 phantom
 * "duplicate H1" reports when this audit was first written.
 */
function linesOutsideFences(text) {
  let inFence = false;
  return text.split('\n').map((line) => {
    if (line.trimStart().startsWith('```')) {
      inFence = !inFence;
      return '';
    }
    return inFence ? '' : line;
  });
}

/** Splits a manual into its H1s and its ordered H2s, ignoring fenced code. */
function headings(text) {
  const lines = linesOutsideFences(text);
  return {
    h1: lines.filter((l) => l.startsWith('# ')),
    h2: lines.filter((l) => l.startsWith('## ')).map((l) => l.slice(3).trim()),
  };
}

/**
 * The body of one H2 section, fences included — null when the section is absent.
 *
 * The section boundary is found on the fence-masked view, then sliced out of
 * the real lines. Searching the raw text instead ended the 起動方法 section of
 * four manuals at a `# または` comment inside their ```sh block, hiding the
 * `go run ./cmd/server` line that came after it and reporting four conforming
 * files as broken.
 */
function sectionBody(text, name) {
  const lines = text.split('\n');
  const masked = linesOutsideFences(text);
  const start = masked.findIndex((l) => l.trim() === `## ${name}`);
  if (start === -1) return null;
  const rest = masked.slice(start + 1);
  const end = rest.findIndex((l) => l.startsWith('## '));
  return lines.slice(start + 1, end === -1 ? lines.length : start + 1 + end).join('\n');
}

/**
 * Every problem with one manual, as stable `class:detail` strings so the
 * report can group by class and the worklist can filter by it.
 */
function auditManual(kind, game, text, navName) {
  const spec = SPEC[kind];
  const issues = [];
  const { h1, h2 } = headings(text);

  const wantTitle = `# ${navName}${spec.suffix}`;
  if (h1.length === 0) issues.push('title-missing');
  else if (h1.length > 1) issues.push(`title-duplicate:${h1.length}`);
  if (h1.length > 0 && h1[0].trim() !== wantTitle) issues.push(`title-format:${h1[0].trim()}`);

  for (const s of spec.sections) {
    const n = h2.filter((x) => x === s).length;
    if (n === 0) issues.push(`section-missing:${s}`);
    else if (n > 1) issues.push(`section-duplicate:${s}`);
  }
  const present = spec.sections.filter((s) => h2.includes(s)).map((s) => h2.indexOf(s));
  if (present.some((v, i) => i > 0 && v < present[i - 1])) issues.push('section-order');

  const mermaid = text.match(/```mermaid\r?\n([\s\S]*?)^```/gm) ?? [];
  if (mermaid.length === 0) issues.push('mermaid-missing');
  else if (!/```mermaid\r?\n\s*flowchart/.test(text)) issues.push('mermaid-not-flowchart');

  // Leftover scaffolding from the template: the placeholder name, the
  // "copy this file" banner, and the instruction comment telling the author
  // to keep API specs out. All three are notes to the author, not content.
  if (/＜ゲーム名＞|<!-- テンプレート:|<!-- API仕様は/.test(text)) issues.push('template-placeholder');

  const launch = sectionBody(text, '起動方法') ?? '';
  if (kind === 'cui') {
    if (!launch.includes(`go run ./cmd/trumpcards ${game}`)) issues.push('launch-cmd-missing');
    const cmds = sectionBody(text, 'コマンド一覧') ?? '';
    if (!/\|\s*コマンド\s*\|\s*短縮形\s*\|\s*説明\s*\|/.test(cmds)) issues.push('command-table-columns');
    // The command may carry arguments -- spider documents `reset [1\|2\|4]` --
    // so match the token, not a closing backtick right after it.
    for (const c of ['reset', 'quit', 'help']) {
      if (!new RegExp(`\\|\\s*\`${c}\\b`).test(cmds)) issues.push(`command-row-missing:${c}`);
    }
  } else {
    if (!launch.includes('go run ./cmd/trumpcards web')) issues.push('launch-cmd-missing');
    if (!launch.includes('go run ./cmd/server')) issues.push('server-cmd-missing');
    // The Web template forbids API specs here — they belong in api/openapi.yaml.
    // Comments are stripped first so the template's own "keep API specs out"
    // note is reported as leftover scaffolding rather than as a leaked spec.
    if (/\bPOST\s+\/|openapi/i.test(text.replace(/<!--[\s\S]*?-->/g, ''))) issues.push('api-spec-leaked');
  }
  return issues;
}

const nav = JSON.parse(readFileSync(NAV_JA, 'utf8')).nav ?? {};
const report = {};
let scanned = 0;

for (const kind of Object.keys(SPEC)) {
  report[kind] = {};
  for (const entry of readdirSync(join(MANUAL_DIR, kind)).sort()) {
    if (!entry.endsWith('.md')) continue;
    const game = entry.slice(0, -3);
    scanned += 1;
    const navName = nav[game];
    if (!navName) {
      report[kind][game] = ['nav-name-missing'];
      continue;
    }
    const issues = auditManual(kind, game, readFileSync(join(MANUAL_DIR, kind, entry), 'utf8'), navName);
    if (issues.length > 0) report[kind][game] = issues;
  }
}

if (scanned < MIN_MANUALS) {
  console.error(
    `audit-manual-template: only ${scanned} manuals scanned, expected at least ${MIN_MANUALS}. ` +
      'The walk or the layout changed — fix this script rather than trusting a clean report.',
  );
  process.exit(1);
}

if (process.argv.includes('--json')) {
  console.log(JSON.stringify(report, null, 1));
  process.exit(0);
}

let total = 0;
for (const kind of Object.keys(SPEC)) {
  const files = Object.keys(report[kind]);
  total += files.length;
  console.log(`\n=== ${kind}: ${files.length} non-conforming manual(s) ===`);
  const byClass = new Map();
  for (const [game, issues] of Object.entries(report[kind])) {
    for (const issue of issues) {
      const cls = issue.startsWith('section-') || issue.startsWith('command-row-') ? issue : issue.split(':')[0];
      if (!byClass.has(cls)) byClass.set(cls, []);
      byClass.get(cls).push(game);
    }
  }
  for (const [cls, games] of [...byClass].sort((a, b) => b[1].length - a[1].length)) {
    console.log(`  ${cls.padEnd(30)} ${String(games.length).padStart(4)}  e.g. ${games.slice(0, 3).join(', ')}`);
  }
}

if (total > 0) {
  console.error(`\naudit-manual-template: ${total} manual(s) of ${scanned} do not follow the template.`);
  process.exit(1);
}
console.log(`\naudit-manual-template: OK (${scanned} manuals all follow the template).`);
