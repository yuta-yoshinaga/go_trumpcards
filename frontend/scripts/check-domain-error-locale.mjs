#!/usr/bin/env bun
// Guard against adding more player-facing Japanese text to domain errors (#7592).

import { readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');
// **internal/ 全体を歩く。** NewDomainError の呼び出しはドメインだけではなく、
// internal/usecase/BouillotteInteractor.go や PrimeroInteractor.go からも呼ばれている。
// いまはどれも英語なので件数は変わらないが、domain だけを見ていると
// そこに日本語を足したときに天井をすり抜ける (PR #7810 のレビュー指摘)。
const SCAN_ROOT = path.join(ROOT, 'internal');

// 日本語リテラルは 0 件。1 件でも増えたら落とす。
const JAPANESE_LITERAL_CEILING = 0;
// 試験は数え方を検査するので、本番の天井 0 を当てるとリテラル 1 件のフィクスチャが違反になって数え方を検査できない。
// 超過検出の試験は 793 件なので 100 で落ちる。
const FIXTURE_CEILING = 100;
const CEILING = process.argv[2] ? FIXTURE_CEILING : JAPANESE_LITERAL_CEILING;
// 実測 4859 ファイル (2026-09-18)。walk が壊れたことを検出するための床であって、違反数の床ではない。
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

function japaneseLiteral(rawArgument) {
  const argument = rawArgument.trim();
  if (argument.startsWith('"') && argument.endsWith('"')) return argument;
  if (!/^fmt\.Sprintf\s*\(/.test(argument)) return null;

  const formatArguments = callArguments(argument, argument.indexOf('('));
  const format = formatArguments[0]?.trim();
  return format?.startsWith('"') && format.endsWith('"') ? format : null;
}

function japaneseNewDomainErrors(source) {
  let count = 0;
  for (const match of source.matchAll(/\bNewDomainError\s*\(/g)) {
    const open = match.index + match[0].lastIndexOf('(');
    const args = callArguments(source, open);
    const literal = args[1] ? japaneseLiteral(args[1]) : null;
    if (literal && /[\u3040-\u309f\u30a0-\u30ff\u3400-\u4dbf\u4e00-\u9fff]/.test(literal)) count += 1;
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
let count = 0;
let matchingFiles = 0;
for (const file of files) {
  const found = japaneseNewDomainErrors(await Bun.file(file).text());
  count += found;
  if (found > 0) matchingFiles += 1;
}

if (!process.argv[2]) {
  assertFloor('domain-error-locale', files.length, SCANNED_FILE_FLOOR, 'Go files scanned');
}

if (count > CEILING) {
  console.error(`domain-error-locale: ${count} Japanese literals in NewDomainError exceeds ceiling ${CEILING}.`);
  process.exit(1);
}
console.log(
  `domain-error-locale: OK (${count} Japanese literals in NewDomainError across ${matchingFiles} file(s); ${files.length} Go files scanned).`,
);
