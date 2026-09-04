#!/usr/bin/env bun
/**
 * Guard: standardize landscapeBanner text across all games.
 *
 * Historically, the landscape banner text for 84 games split into 8 variations
 * in Japanese and 9 in English, causing the UI presentation to flutter when
 * switching games. This guard ensures the banner text remains exactly two variants
 * per language: the standard text, and the specific text for `cruel`.
 *
 * `cruel` is the only exception: its text mentions the 12-column tableau, which
 * provides game-specific context that would be lost if standardized.
 */
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { assertFloor } from './lib/floor.mjs';

const LOCALES = new URL('../src/i18n/locales/', import.meta.url).pathname;

// Number of games, not entries (since both ja and en are checked)
let gamesInspected = 0;
let failed = false;

for (const lang of ['ja', 'en']) {
  const dir = join(LOCALES, lang);
  const files = readdirSync(dir).filter((f) => f.endsWith('.json'));

  const variants = new Map();
  let langInspected = 0;

  for (const file of files) {
    const game = file.replace('.json', '');
    const content = JSON.parse(readFileSync(join(dir, file), 'utf-8'));

    if ('landscapeBanner' in content) {
      const text = content.landscapeBanner;
      if (!variants.has(text)) {
        variants.set(text, []);
      }
      variants.get(text).push(game);
      langInspected++;
    }
  }

  if (lang === 'ja') {
    gamesInspected = langInspected;
  }

  // We expect exactly 2 distinct banner texts per language (the standard one + cruel's exception)
  // Or rather, the prompt asks to check that it is *exactly* 2.
  if (variants.size !== 2) {
    console.error(`\n[${lang}] Found ${variants.size} variations of landscapeBanner (expected exactly 2):`);
    for (const [text, games] of variants.entries()) {
      console.error(`  "${text}":`);
      // print games nicely
      console.error(`    ${games.join(', ')}`);
    }
    failed = true;
  }
}

assertFloor('landscape-banner', gamesInspected, 60, 'games');

if (failed) {
  process.exit(1);
}

console.log(`checked landscapeBanner across ${gamesInspected} games`);
