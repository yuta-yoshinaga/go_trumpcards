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

const readJson = (file) => JSON.parse(readFileSync(file, 'utf8'));
const cuiFiles = (language) => readdirSync(join(CUI_LOCALES, language)).filter((file) => file.endsWith('.json'));

const mismatches = [];
let compared = 0;

for (const language of LANGUAGES) {
  const messageCode = readJson(join(WEB_LOCALES, language, 'common.json')).messageCode;
  for (const file of cuiFiles(language)) {
    const game = file.slice(0, -'.json'.length);
    const cui = readJson(join(CUI_LOCALES, language, file));
    for (const [key, cuiText] of Object.entries(cui)) {
      if (!ERROR_KEY.test(key)) continue;
      const webKey = `${game}.${key}`;
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
    }
  }
}

if (IS_FIXTURE) {
  assertFloor('locale-text-agreement', compared, 1, 'shared error messages compared');
} else {
  assertFloor('locale-text-agreement', compared, 577, 'shared error messages compared');
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

console.log(`locale-text-agreement: OK (${compared} shared error messages compared).`);
