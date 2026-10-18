import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * Every hand card stays reachable without scrolling the footer at 375×667 (#11655) and 1280×800 (#11701).
 *
 * Since #11611 pins `actions` to the bottom of `GameFooter`, the rest of the footer
 * (`game-footer-content`) is a nested scroller inside the mobile cap, and a hand placed there
 * can be cut off. #11655 moved everything but the hand out of it, stacked large hands
 * (`stackRows`) and gave the largest a taller cap (`tall`). The rule it settled on: the top
 * 52px of every card in the footer content — the rank corner plus a 44px tap target —
 * must lie inside the footer content's visible box. footer-budget.spec.ts measures scroll
 * overflow instead, which overstates the problem once only the hand is left.
 *
 * Hand sizes depend on the shuffled deal (#4373), so one load per page is intentional; the
 * pages were checked at their worst of 4–10 deals when this guard was added. Only the bottom
 * edge is checked; horizontal overflow is out of scope (MobileHandGrid fits rows to the width).
 * A failure means "a deal or a layout change hides a card", never "retry".
 * Routes come from gameRoutes.ts (#11821). The original fixed list remains a strict floor for
 * pages known to put hands in the footer, so losing their footer or cards still fails. Other
 * routes are measured when they have footer hands and annotated as skipped when they do not;
 * this two-tier guard covers newly added footer hands without treating play-area hands as defects.
 */
/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;

/**
 * Pages whose footer content overflowed at the start of #11655 (and the very large hands), minus
 * the ones whose hand lives in the play area (Crazy Four Poker, Double Attack, Double Exposure,
 * Free Bet, Michigan, Spanish 21, Trente et Quarante). These routes must keep their footer hand.
 */
const FOOTER_HAND_PAGES: readonly string[] = [
  '/badugi',
  '/batak',
  '/bauernschnapsen',
  '/bigo',
  '/bigohilo',
  '/binokel',
  '/biriba',
  '/bolivia',
  '/boston',
  '/bridge',
  '/burraco',
  '/callbreak',
  '/canasta',
  '/chicago',
  '/chineseten',
  '/cinch',
  '/coinche',
  '/continentalrummy',
  '/courchevel',
  '/courchevelhilo',
  '/crazypineapple',
  '/daifugo',
  '/deucetoseven',
  '/doubt',
  '/dramaha',
  '/fivecardstud',
  '/followthequeen',
  '/fortyfives',
  '/frenchtarot',
  '/ganjifa',
  '/guandan',
  '/guts',
  '/handandfoot',
  '/holdem',
  '/honeymoonbridge',
  '/indianpoker',
  '/indianrummy',
  '/irishpoker',
  '/israeliwhist',
  '/kaiser',
  '/king',
  '/literature',
  '/lingerlonger',
  '/loo',
  '/machiavelli',
  '/marriage',
  '/mighty',
  '/minchiate',
  '/mushi',
  '/mus',
  '/nap',
  '/ohhell',
  '/omaha',
  '/omahahilo',
  '/ombre',
  '/piedmontesetarot',
  '/pig',
  '/pineapple',
  '/poker',
  '/popejoan',
  '/quadrille',
  '/razz',
  '/rummy500',
  '/reversis',
  '/samba',
  '/scarto',
  '/schafkopf',
  '/scopone',
  '/sevencardstud',
  '/sevencardstudhilo',
  '/sevens',
  '/sergeantmajor',
  '/shengji',
  '/shortdeck',
  '/sixbidsolo',
  '/soko',
  '/spades',
  '/stealingbundles',
  '/tarneeb',
  '/tarocchini',
  '/teendopaanch',
  '/tongits',
  '/ulti',
  '/vint',
  '/vira',
  '/watten',
  '/zwicker',
  '/shelem',
  '/tarabish',
  '/rams',
];

const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

test('route list is complete', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(300);
});

test('every strict footer-hand page is a game route', () => {
  const missingPages = FOOTER_HAND_PAGES.filter((path) => !routes.includes(path));
  expect(missingPages, `footer-hand pages missing from game routes: ${missingPages.join(', ')}`).toEqual([]);
});

/** Enforces the 52px visibility rule, with strict presence floors for known footer-hand pages. */
async function checkHandVisibility(page: import('@playwright/test').Page, path: string) {
  await navigateTo(page, path);
  // Cards deal in with a staggered slide-up animation; measure their resting positions, not
  // the last cards still sliding into place. A not-yet-started (pending) animation counts as
  // unfinished; infinite ones (e.g. a pulsing highlight) never finish, so they are not awaited.
  await page.waitForFunction(() =>
    document
      .getAnimations()
      .every(
        (a) =>
          a.playState === 'finished' ||
          a.playState === 'idle' ||
          a.effect?.getComputedTiming().iterations === Number.POSITIVE_INFINITY,
      ),
  );
  const result = await page.evaluate((visiblePx) => {
    const footer = document.querySelector('[data-testid="game-footer-content"]');
    if (!footer) return { found: false, cards: 0, hidden: [] as number[] };
    const bottom = footer.getBoundingClientRect().bottom;
    // Cards are <img> (CardImage) or <div role="img"> (CardFace, e.g. tarot decks).
    const boxes = [...footer.querySelectorAll('img, [role="img"]')]
      .map((el) => el.getBoundingClientRect())
      .filter((box) => box.width >= 20 && box.height >= 30);
    return {
      found: true,
      cards: boxes.length,
      hidden: boxes.filter((box) => box.top + visiblePx > bottom).map((box) => Math.round(box.top)),
    };
  }, VISIBLE_PX);
  const requiresFooterHand = FOOTER_HAND_PAGES.includes(path);
  if (!result.found) {
    if (requiresFooterHand) expect(result.found, `${path}: game-footer-content not found`).toBe(true);
    test.info().annotations.push({
      type: 'skipped page',
      description: 'game-footer-content was not present on this route.',
    });
    return;
  }
  if (result.cards === 0) {
    if (requiresFooterHand) {
      expect(result.cards, `${path}: no hand cards found in the footer content`).toBeGreaterThan(0);
    }
    test.info().annotations.push({
      type: 'skipped page',
      description: 'No hand cards were present in game-footer-content.',
    });
    return;
  }
  expect(
    result.hidden,
    `${path}: cards whose top ${VISIBLE_PX}px is below the footer content (tops: ${result.hidden.join(', ')})`,
  ).toEqual([]);
}

test.describe('phone 375x667', () => {
  test.use({ viewport: { width: 375, height: 667 } });
  for (const path of routes) {
    test(`${path} keeps every footer hand card reachable on a phone`, async ({ page }) => {
      await checkHandVisibility(page, path);
    });
  }
});

test.describe('desktop 1280x800', () => {
  test.use({ viewport: { width: 1280, height: 800 } });
  for (const path of routes) {
    test(`${path} keeps every footer hand card reachable on a desktop`, async ({ page }) => {
      await checkHandVisibility(page, path);
    });
  }
});
