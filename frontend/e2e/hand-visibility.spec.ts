import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * Every hand card stays reachable without scrolling the footer at 375×667 (#11655).
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
 */
test.use({ viewport: { width: 375, height: 667 } });

/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;

/**
 * Pages whose footer content overflowed at the start of #11655 (and the very large hands), minus
 * the ones whose hand lives in the play area rather than the footer (Crazy Four Poker, Double
 * Attack, Double Exposure, Free Bet, Michigan, Spanish 21, Trente et Quarante).
 */
const PAGES: readonly string[] = [
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
  '/indianpoker',
  '/indianrummy',
  '/irishpoker',
  '/kaiser',
  '/king',
  '/literature',
  '/loo',
  '/machiavelli',
  '/marriage',
  '/mighty',
  '/minchiate',
  '/mus',
  '/nap',
  '/ohhell',
  '/omaha',
  '/omahahilo',
  '/ombre',
  '/piedmontesetarot',
  '/pineapple',
  '/poker',
  '/quadrille',
  '/razz',
  '/rummy500',
  '/samba',
  '/scarto',
  '/schafkopf',
  '/sevencardstud',
  '/sevencardstudhilo',
  '/sevens',
  '/shengji',
  '/shortdeck',
  '/sixbidsolo',
  '/soko',
  '/spades',
  '/tarneeb',
  '/tarocchini',
  '/tongits',
  '/ulti',
  '/vint',
  '/vira',
  '/watten',
];

for (const path of PAGES) {
  test(`${path} keeps every footer hand card reachable on a phone`, async ({ page }) => {
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
    // Floors: a missing footer or a selector that matches no card must fail, not pass vacuously.
    expect(result.found, `${path}: game-footer-content not found`).toBe(true);
    expect(result.cards, `${path}: no hand cards found in the footer content`).toBeGreaterThan(0);
    expect(
      result.hidden,
      `${path}: cards whose top ${VISIBLE_PX}px is below the footer content (tops: ${result.hidden.join(', ')})`,
    ).toEqual([]);
  });
}
