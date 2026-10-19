import { expect, type Page, test } from '@playwright/test';
import { waitForAnimations } from '../helpers';

/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;

/**
 * Pages whose footer content overflowed at the start of #11655 (and the very large hands), minus
 * the ones whose hand lives in the play area (Crazy Four Poker, Double Attack, Double Exposure,
 * Free Bet, Michigan, Spanish 21, Trente et Quarante). These routes must keep their footer hand.
 */
export const FOOTER_HAND_PAGES: readonly string[] = [
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

/** Enforces the 52px visibility rule, with strict presence floors for known footer-hand pages. */
export async function checkHand(page: Page, path: string, _viewport: { width: number; height: number }): Promise<void> {
  // Cards deal in with a staggered slide-up animation; measure their resting positions, not
  // the last cards still sliding into place.
  let settled = true;
  try {
    await waitForAnimations(page);
  } catch {
    settled = false;
  }
  expect.soft(settled, `Hand check: ${path}: animations did not settle`).toBe(true);
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
    if (requiresFooterHand) expect.soft(result.found, `Hand check: ${path}: game-footer-content not found`).toBe(true);
    test.info().annotations.push({
      type: 'skipped page',
      description: 'game-footer-content was not present on this route.',
    });
    return;
  }
  if (result.cards === 0) {
    if (requiresFooterHand) {
      expect.soft(result.cards, `Hand check: ${path}: no hand cards found in the footer content`).toBeGreaterThan(0);
    }
    test.info().annotations.push({
      type: 'skipped page',
      description: 'No hand cards were present in game-footer-content.',
    });
    return;
  }
  expect
    .soft(
      result.hidden,
      `Hand check: ${path}: cards whose top ${VISIBLE_PX}px is below the footer content (tops: ${result.hidden.join(', ')})`,
    )
    .toEqual([]);
}
