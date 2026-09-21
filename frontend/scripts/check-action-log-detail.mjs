#!/usr/bin/env bun
// Guard against adding more literal action-log details during ADR-0040 migration.

import { readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');
const SCAN_ROOT = path.join(ROOT, 'internal');

// **ADR-0040 の移行は完了した (2338 件 -> 0 件)。この天井は 0 のまま動かさない。**
// 0 では「違反が無い」と「走査していない」が同じ出力になるので、意味のある床は
// 下の SCANNED_FILE_FLOOR (走査した Go ファイル数) の方に移してある
// —— `check-domain-error-locale.mjs` が #7592 の完了時に取ったのと同じ形。
const DETAIL_LITERAL_CEILING = 0;
// フィクスチャは数え方のテスト用で本番の件数ではないため、天井を緩くする。
const FIXTURE_CEILING = 100;
const CEILING = process.argv[2] ? FIXTURE_CEILING : DETAIL_LITERAL_CEILING;
// 天井が 0 になった今、走査が実際に行われたことを保証するのはこの床だけ。
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
    (argumentCount >= 3 && !isLiteralDetailCode(argument) && !/[()]/.test(argument))
  );
}

function isDetailParams(rawArgument) {
  const argument = rawArgument.trim();
  return argument === 'nil' || /^map\s*\[\s*string\s*\]\s*string\s*\{/.test(argument);
}

function isLiteralDetailCode(rawArgument) {
  const argument = rawArgument.trim();
  return argument.startsWith('"') || /^fmt\.Sprintf\s*\(/.test(argument);
}

function detailCodeText(rawArgument) {
  return rawArgument.trim() || '<missing>';
}

function isValidDetailCode(rawArgument) {
  return /^"[a-z0-9]+\.log\.[A-Za-z0-9_.]+"$/.test(rawArgument.trim());
}

function isFunctionDeclaration(source, index) {
  const lineStart = source.lastIndexOf('\n', index - 1) + 1;
  const prefix = source.slice(lineStart, index);
  return /\bfunc\b/.test(prefix) && !prefix.includes('{');
}

function literalActionLogs(source, file) {
  const violations = [];
  for (const match of source.matchAll(/\b(?:appendLog|addLog)[A-Za-z0-9_]*\s*\(/g)) {
    if (isFunctionDeclaration(source, match.index)) continue;
    const open = match.index + match[0].lastIndexOf('(');
    const args = callArguments(source, open);
    if (args.length === 0) continue;
    const detailArgs = isCardsArgument(args.at(-1), args.length) ? args.slice(0, -1) : args;
    const detailParams = detailArgs.at(-1);
    const detailCode = detailArgs.at(-2);
    if (!isDetailParams(detailParams ?? '')) {
      if (isLiteralDetailCode(detailCode ?? '') && !isValidDetailCode(detailCode ?? '')) {
        violations.push({ file, detailCode: detailCodeText(detailCode ?? '') });
      }
      continue;
    }
    if (isLiteralDetailCode(detailCode ?? '') && !isValidDetailCode(detailCode ?? '')) {
      violations.push({ file, detailCode: detailCodeText(detailCode ?? '') });
    }
  }
  return violations;
}

let files;
try {
  files = await goFiles(SCAN_ROOT);
} catch (error) {
  if (error?.code !== 'ENOENT') throw error;
  files = [];
}

if (!process.argv[2]) assertFloor('action-log-detail', files.length, SCANNED_FILE_FLOOR, 'Go files scanned');

let matchingFiles = 0;
const violations = [];
for (const file of files) {
  const found = literalActionLogs(await Bun.file(file).text(), path.relative(ROOT, file));
  violations.push(...found);
  if (found.length > 0) matchingFiles += 1;
}
const count = violations.length;

if (count > CEILING) {
  for (const violation of violations) {
    console.error(`action-log-detail: ${violation.file}: invalid detailCode ${violation.detailCode}`);
  }
  console.error(`action-log-detail: ${count} literal details exceeds ceiling ${CEILING}.`);
  process.exit(1);
}
if (count < CEILING)
  console.warn(`action-log-detail: ${count} literal details; lower DETAIL_LITERAL_CEILING from ${CEILING}.`);
console.log(
  `action-log-detail: OK (${count} literal details across ${matchingFiles} file(s); ${files.length} Go files scanned).`,
);
