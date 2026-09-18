#!/usr/bin/env bun
// Guard that shared CUI and Web error messages have identical text.

import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const IS_FIXTURE = Boolean(process.argv[2]);
const FIXTURE_ROOT = IS_FIXTURE ? resolve(process.argv[2]) : resolve(new URL('../..', import.meta.url).pathname);
const CUI_LOCALES = join(FIXTURE_ROOT, 'internal', 'i18n', 'locales');
const WEB_LOCALES = join(FIXTURE_ROOT, 'frontend', 'src', 'i18n', 'locales');
const LANGUAGES = ['en', 'ja'];
const ERROR_KEY = /^err[A-Z]/;
const LOG_KEY = /^log\./;

const readJson = (file) => JSON.parse(readFileSync(file, 'utf8'));
const cuiFiles = (language) => readdirSync(join(CUI_LOCALES, language)).filter((file) => file.endsWith('.json'));

const mismatches = [];
let compared = 0;
let logCompared = 0;

for (const language of LANGUAGES) {
  const common = readJson(join(WEB_LOCALES, language, 'common.json'));
  const messageCode = common.messageCode;
  for (const file of cuiFiles(language)) {
    const game = file.slice(0, -'.json'.length);
    const cui = readJson(join(CUI_LOCALES, language, file));
    for (const [key, cuiText] of Object.entries(cui)) {
      const webKey = `${game}.${key}`;
      if (ERROR_KEY.test(key)) {
        if (!(webKey in messageCode)) continue;
        compared += 1;
        if (cuiText !== messageCode[webKey]) {
          mismatches.push({
            language,
            game,
            key,
            cuiText,
            webText: messageCode[webKey],
          });
        }
      } else if (LOG_KEY.test(key)) {
        if (!(webKey in common)) {
          mismatches.push({ language, game, key, cuiText, webText: undefined });
          continue;
        }
        logCompared += 1;
        if (cuiText !== common[webKey]) {
          mismatches.push({
            language,
            game,
            key,
            cuiText,
            webText: common[webKey],
          });
        }
      }
    }
  }
}

if (IS_FIXTURE) {
  assertFloor('locale-text-agreement', compared, 1, 'shared error messages compared');
  // log 系はフィクスチャに床を張らない。「Web にキーが無い」ことを検査するフィクスチャは
  // 比較 0 件が正しい状態で、床を張るとその試験自体が床で落ちる。
} else {
  assertFloor('locale-text-agreement', compared, 577, 'shared error messages compared');
  // 実測 32 (2026-09-18、移行済み 2 ゲーム)。移行が進むほど増えるので下がることはない
  assertFloor('locale-text-agreement', logCompared, 20, 'shared log messages compared');
}

if (mismatches.length > 0) {
  console.error(`locale-text-agreement: ${mismatches.length} mismatch(es).`);
  for (const { language, game, key, cuiText, webText } of mismatches) {
    console.error(`  ${language} ${game}.${key}`);
    console.error(`    CUI: ${cuiText}`);
    console.error(`    Web: ${webText}`);
  }
  process.exit(1);
}

console.log(
  `locale-text-agreement: OK (${compared} shared error messages, ${logCompared} shared log messages compared).`,
);
