#!/usr/bin/env bun
/**
 * Guard that Japanese locale values are translated and do not retain English labels.
 *
 * This checks the CUI locale tree because it is the source of the strings that exposed
 * #7962.  The optional root argument lets the test spawn this exact script against a
 * temporary fixture tree.
 */
import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const IS_FIXTURE = Boolean(process.argv[2]);
const ROOT = IS_FIXTURE ? resolve(process.argv[2]) : resolve(new URL('../..', import.meta.url).pathname);
const LOCALES = join(ROOT, 'internal', 'i18n', 'locales');
const JAPANESE = /[\u3040-\u30ff\u3400-\u9fff]/;
const ENGLISH_WORD = /\b[A-Za-z]{3,}\b/;
const PLACEHOLDER = /{{[^{}]*}}/g;
const FILES = readdirSync(join(LOCALES, 'ja'))
  .filter((file) => file.endsWith('.json'))
  .sort();

const read = (language, file) => JSON.parse(readFileSync(join(LOCALES, language, file), 'utf8'));
const withoutPlaceholders = (value) => value.replace(PLACEHOLDER, '');

/** Flatten locale objects to the leaf keys used by the CUI locale files. */
function leaves(node, prefix = '', out = new Map()) {
  if (node !== null && typeof node === 'object' && !Array.isArray(node)) {
    for (const [key, value] of Object.entries(node)) leaves(value, prefix ? `${prefix}.${key}` : key, out);
  } else {
    out.set(prefix, node);
  }
  return out;
}

// A: command syntax such as `m t <col> t <col>` is intentionally identical in ja/en.
// The tokens after a label here (`type: 1=スート`, `nr / nextround: 次のラウンドへ`) are the
// literal words the player types, so they are exempt under BOTH conditions.
// Measured exemption count: 279 for condition 1, 7 for condition 2.
// `usage` is listed for condition 2's sake: it exempts 0 additional keys under condition 1.
const isHelpSyntax = (key) => /^(help|prompt|hintReason|usage)/.test(key);
// B: angle-bracket arguments and aligned tables are formatting, not prose.
// Measured exemption count: 271 (some are also covered by A).
const isCommandOrTable = (value) => value.includes('<') || / {2,}/.test(value);
// C: `--- PLAYER ---` separators are monospaced headings surrounded by border characters.
// Measured exemption count: 37.
const isHeader = (key) => /Header$/.test(key);
// D: `CPU {{idx}}` follows the repository-wide spelling convention.
// Measured exemption count: 26 under condition 1. Condition 2 is covered by the
// `CPU` entry in ESTABLISHED_TERMS when the value contains that established term.
const isCpuKey = (key) => /cpu/i.test(key);

// E: these are established names or spoken terms, so translating their spelling would be
// incorrect. Each entry has its own reason:
const ESTABLISHED_TERMS = [
  'CPU', // Seat label shared by every game.
  'VPIP', // Standard poker statistic, conventionally written this way in Japanese sources.
  'PFR', // Standard poker statistic, conventionally written this way in Japanese sources.
  '3Bet', // Standard poker statistic, conventionally written this way in Japanese sources.
  'AF', // Standard poker statistic, conventionally written this way in Japanese sources.
  'Hi-Lo', // Counting-method name, like its sibling KO / Zen Count / Omega II names.
  'Zen', // Counting-method name, like its sibling KO / Zen Count / Omega II names.
  'Omega', // Counting-method name, like its sibling KO / Zen Count / Omega II names.
  'Tichu', // Declaration name; this is the word actually spoken at the table.
  'Go', // Cribbage declaration; helpGo explicitly says to declare Go.
  'Capot', // Quansh/Polignac role name; the existing カポー translation has zero occurrences.
  'capot', // Same role name in the spelling used by a lowercase label.
  'Queens Up', // Crazy 4 Poker bet name; the existing クイーンズ translation has zero occurrences.
  'Weis', // Jass term written in Latin script throughout jass.json, alongside Stöck and Schieber.
];
const hasEstablishedTerm = (value) => ESTABLISHED_TERMS.some((term) => value.includes(term));

// G: these are not labels: one is an executable name and the other is a URL scheme.
// Measured exemption count: 1 each.
const isExecutableName = (value) => value.includes('trumpcards');
const isUrl = (value) => value.includes('://');

// F: outputTitle is the game’s formal name. This is deliberately a key-based exemption:
// it covers the ten measured formal-name cases, but also lets a different game name pass
// (for example, the historical soko = Five Card Stud mistake from #7974).
const isFormalGameTitle = (key) => /^outputTitle/.test(key);

const condition1Violations = [];
const condition2Violations = [];
let identicalCompared = 0;
let identicalExempt = 0;
let englishLabelCompared = 0;
let englishLabelExempt = 0;

for (const file of FILES) {
  const ja = leaves(read('ja', file));
  const en = leaves(read('en', file));
  for (const [key, value] of ja) {
    if (typeof value !== 'string' || value.length === 0) continue;

    const englishValue = en.get(key);
    if (typeof englishValue === 'string') identicalCompared += 1;

    const plain = withoutPlaceholders(value);
    if (!JAPANESE.test(value) && value === englishValue && ENGLISH_WORD.test(plain)) {
      const exempt =
        isHelpSyntax(key) ||
        isCommandOrTable(value) ||
        isHeader(key) ||
        isCpuKey(key) ||
        hasEstablishedTerm(value) ||
        isFormalGameTitle(key);
      if (exempt) identicalExempt += 1;
      else condition1Violations.push(`${file}:${key} = ${value}`);
    }

    if (!JAPANESE.test(value)) continue;
    englishLabelCompared += 1;
    const label = /(?:^|[|\s])([A-Za-z][A-Za-z]{2,})\s*:/.test(plain) || /\b([a-z][a-zA-Z]{2,})=/.test(plain);
    if (!label) continue;
    const exempt =
      file.startsWith('cli_') ||
      /^opt[A-Z]/.test(key) ||
      isHelpSyntax(key) ||
      hasEstablishedTerm(value) ||
      isExecutableName(value) ||
      isUrl(value);
    if (exempt) englishLabelExempt += 1;
    else condition2Violations.push(`${file}:${key} = ${value}`);
  }
}

if (IS_FIXTURE) {
  assertFloor('japanese-locale-translated', identicalCompared, 1, 'keys compared for identical-to-en');
  assertFloor('japanese-locale-translated', englishLabelCompared, 1, 'Japanese keys checked for English labels');
} else {
  assertFloor('japanese-locale-translated', identicalCompared, 12000, 'keys compared for identical-to-en');
  assertFloor('japanese-locale-translated', englishLabelCompared, 11500, 'Japanese keys checked for English labels');
}

if (condition1Violations.length > 0 || condition2Violations.length > 0) {
  if (condition1Violations.length > 0) {
    console.error(`japanese-locale-translated: identical-to-en: ${condition1Violations.length} violation(s).`);
    for (const violation of condition1Violations) console.error(`  ${violation}`);
  }
  if (condition2Violations.length > 0) {
    console.error(
      `japanese-locale-translated: english-label-in-japanese: ${condition2Violations.length} violation(s).`,
    );
    for (const violation of condition2Violations) console.error(`  ${violation}`);
  }
  process.exit(1);
}

console.log(
  `japanese-locale-translated: OK (${identicalCompared} keys compared; identical-to-en: 0 violations, ${identicalExempt} exempt; ` +
    `english-label-in-japanese: ${englishLabelCompared} keys checked, 0 violations, ${englishLabelExempt} exempt).`,
);
