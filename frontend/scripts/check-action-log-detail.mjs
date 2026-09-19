#!/usr/bin/env bun
// Guard against adding more literal action-log details during ADR-0040 migration.

import { readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');
const SCAN_ROOT = path.join(ROOT, 'internal');

// ADR-0040 の移行が進むたびに実測して下げる。0 になったら
// `check-domain-error-locale.mjs` と同じく床を走査ファイル数へ移す。
// 実測値: 実測 1436 件。
const DETAIL_LITERAL_CEILING = 214;
// フィクスチャは数え方のテスト用で本番の件数ではないため、天井を緩くする。
const FIXTURE_CEILING = 100;
const CEILING = process.argv[2] ? FIXTURE_CEILING : DETAIL_LITERAL_CEILING;
const SCANNED_FILE_FLOOR = 3000;

async function goFiles(dir) {
  const files = [];
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) files.push(...(await goFiles(full)));
    else if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) files.push(full);
  }
  return files;
}

function skipLiteral(source, start) {
  const quote = source[start];
  let i = start + 1;
  while (i < source.length) {
    if (quote !== '`' && source[i] === '\\') {
      i += 2;
      continue;
    }
    if (source[i] === quote) return i + 1;
    i += 1;
  }
  return source.length;
}

function callArguments(source, open) {
  const args = [];
  let start = open + 1;
  let parens = 0;
  let brackets = 0;
  let braces = 0;
  for (let i = start; i < source.length; i += 1) {
    const ch = source[i];
    if (ch === '"' || ch === '`' || ch === "'") {
      i = skipLiteral(source, i) - 1;
      continue;
    }
    if (ch === '/' && source[i + 1] === '/') {
      const end = source.indexOf('\n', i + 2);
      i = end === -1 ? source.length : end;
      continue;
    }
    if (ch === '/' && source[i + 1] === '*') {
      const end = source.indexOf('*/', i + 2);
      i = end === -1 ? source.length : end + 1;
      continue;
    }
    if (ch === '(') parens += 1;
    else if (ch === ')') {
      if (parens === 0 && brackets === 0 && braces === 0) {
        args.push(source.slice(start, i));
        return args;
      }
      parens -= 1;
    } else if (ch === '[') brackets += 1;
    else if (ch === ']') brackets -= 1;
    else if (ch === '{') braces += 1;
    else if (ch === '}') braces -= 1;
    else if (ch === ',' && parens === 0 && brackets === 0 && braces === 0) {
      args.push(source.slice(start, i));
      start = i + 1;
    }
  }
  return [];
}

function isCardsArgument(rawArgument, argumentCount) {
  const argument = rawArgument.trim();
  return (
    argument === 'nil' ||
    argument.startsWith('[]*Card{') ||
    argument.startsWith('append([]*Card') ||
    /^[A-Za-z_][A-Za-z0-9_.]*(\[[^\]]*\])?$/.test(argument) ||
    (argumentCount >= 3 && !hasLiteralDetail(argument) && !/[()]/.test(argument))
  );
}

function hasLiteralDetail(rawArgument) {
  const argument = rawArgument.trim();
  return argument.startsWith('"') || /^fmt\.Sprintf\s*\(/.test(argument);
}

function literalActionLogs(source) {
  let count = 0;
  for (const match of source.matchAll(/\bappendLog[A-Za-z0-9_]*\s*\(/g)) {
    const open = match.index + match[0].lastIndexOf('(');
    const args = callArguments(source, open);
    if (args.length === 0) continue;
    const detailArgs = isCardsArgument(args.at(-1), args.length) ? args.slice(0, -1) : args;
    if (hasLiteralDetail(detailArgs.at(-1) ?? '')) count += 1;
  }
  return count;
}

let files;
try {
  files = await goFiles(SCAN_ROOT);
} catch (error) {
  if (error?.code !== 'ENOENT') throw error;
  files = [];
}

if (!process.argv[2]) assertFloor('action-log-detail', files.length, SCANNED_FILE_FLOOR, 'Go files scanned');

let count = 0;
let matchingFiles = 0;
for (const file of files) {
  const found = literalActionLogs(await Bun.file(file).text());
  count += found;
  if (found > 0) matchingFiles += 1;
}

if (count > CEILING) {
  console.error(`action-log-detail: ${count} literal details exceeds ceiling ${CEILING}.`);
  process.exit(1);
}
if (count < CEILING)
  console.warn(`action-log-detail: ${count} literal details; lower DETAIL_LITERAL_CEILING from ${CEILING}.`);
console.log(
  `action-log-detail: OK (${count} literal details across ${matchingFiles} file(s); ${files.length} Go files scanned).`,
);
