#!/usr/bin/env bun
// Guard against mismatched action-log placeholder names during ADR-0040 migration.

import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = process.argv[2] ? path.resolve(process.argv[2]) : path.resolve(HERE, '..', '..');
const INTERNAL = path.join(ROOT, 'internal');
const DOMAIN = path.join(INTERNAL, 'domain');
const LOCALES = path.join(INTERNAL, 'i18n', 'locales');
// 実測64コード。移行が進むほど増えるため、下がらない床にする。
const CODES_FLOOR = 40;
// 実測1839ファイル。走査対象が欠落していないことを検査する床にする。
const SCANNED_FILE_FLOOR = 1200;

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

function mapKeys(argument) {
  const map = argument.trim();
  if (map === 'nil') return [];
  if (!map.startsWith('map[string]string{')) return [];
  return [...map.matchAll(/(?:[{,])\s*"([^"\\]+)"\s*:/g)].map((match) => match[1]);
}

function codeCalls(source) {
  const calls = new Map();
  for (const match of source.matchAll(/\bappendLog[A-Za-z0-9_]*\s*\(/g)) {
    const open = match.index + match[0].lastIndexOf('(');
    const args = callArguments(source, open);
    for (let i = 0; i < args.length - 1; i += 1) {
      const code = args[i].trim().match(/^"([A-Za-z0-9_-]+\.log\.[A-Za-z0-9_]+)"$/)?.[1];
      if (!code) continue;
      const keys = calls.get(code) ?? new Set();
      for (const key of mapKeys(args[i + 1])) keys.add(key);
      calls.set(code, keys);
      break;
    }
  }
  return calls;
}

function placeholders(value) {
  return new Set([...value.matchAll(/{{\s*([A-Za-z_][A-Za-z0-9_]*)\s*}}/g)].map((match) => match[1]));
}

function difference(left, right) {
  return [...left].filter((value) => !right.has(value)).sort();
}

let files;
try {
  files = await goFiles(DOMAIN);
} catch (error) {
  if (error?.code !== 'ENOENT') throw error;
  files = [];
}
const calls = new Map();
for (const file of files) {
  for (const [code, keys] of codeCalls(await readFile(file, 'utf8'))) {
    const allKeys = calls.get(code) ?? new Set();
    for (const key of keys) allKeys.add(key);
    calls.set(code, allKeys);
  }
}

const mismatches = [];
const warnings = [];
for (const [code, keys] of [...calls.entries()].sort()) {
  const [game] = code.split('.');
  const localeKey = code.slice(game.length + 1);
  const localePlaceholders = new Map();
  for (const language of ['ja', 'en']) {
    const file = path.join(LOCALES, language, `${game}.json`);
    let locale;
    try {
      locale = JSON.parse(await readFile(file, 'utf8'));
    } catch {
      mismatches.push(`${code}: missing ${language} locale file`);
      localePlaceholders.set(language, null);
      continue;
    }
    if (!(localeKey in locale)) {
      mismatches.push(`${code}: missing ${language} locale key`);
      localePlaceholders.set(language, null);
      continue;
    }
    localePlaceholders.set(language, placeholders(locale[localeKey]));
  }

  const ja = localePlaceholders.get('ja');
  const en = localePlaceholders.get('en');
  if (ja && en) {
    const missingJa = difference(ja, keys);
    const missingEn = difference(en, keys);
    if (missingJa.length) mismatches.push(`${code}: ja placeholders missing from params: ${missingJa.join(', ')}`);
    if (missingEn.length) mismatches.push(`${code}: en placeholders missing from params: ${missingEn.join(', ')}`);

    const unusedJa = difference(keys, ja);
    const unusedEn = difference(keys, en);
    for (const key of unusedJa.filter((value) => unusedEn.includes(value))) {
      mismatches.push(`${code}: param key unused in ja and en: ${key}`);
    }
    for (const key of unusedJa.filter((value) => !unusedEn.includes(value))) {
      warnings.push(`${code}: param key unused in ja: ${key}`);
    }
    for (const key of unusedEn.filter((value) => !unusedJa.includes(value))) {
      warnings.push(`${code}: param key unused in en: ${key}`);
    }
  }
}

for (const warning of warnings) console.warn(`action-log-params: warning: ${warning}`);
for (const mismatch of mismatches) console.error(`action-log-params: mismatch: ${mismatch}`);

if (!process.argv[2]) {
  assertFloor('action-log-params', calls.size, CODES_FLOOR, 'codes checked');
  assertFloor('action-log-params', files.length, SCANNED_FILE_FLOOR, 'Go files scanned');
}
console.log(`action-log-params: checked ${calls.size} codes (${files.length} Go files scanned).`);
if (mismatches.length > 0) process.exit(1);
console.log(`action-log-params: OK (${calls.size} codes checked; ${files.length} Go files scanned).`);
