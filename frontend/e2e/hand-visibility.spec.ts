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
 * pages were checked at their worst of 4–8 deals when this guard was added. A failure means
 * "a deal or a layout change hides a card", never "retry".
 */
test.use({ viewport: { width: 375, height: 667 } });

/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;

/** Pages whose footer content overflowed at the start of #11655 (and the very large hands). */
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
  '/crazyfourpoker',
  '/crazypineapple',
  '/daifugo',
  '/deucetoseven',
  '/doubleattack',
  '/doubleexposure',
  '/doubt',
  '/dramaha',
  '/fivecardstud',
  '/followthequeen',
  '/fortyfives',
  '/freebet',
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
  '/michigan',
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
  '/spanish21',
  '/tarneeb',
  '/tarocchini',
  '/tongits',
  '/trenteetquarante',
  '/ulti',
  '/vint',
  '/vira',
  '/watten',
];

for (const path of PAGES) {
  test(`${path} keeps every footer hand card reachable on a phone`, async ({ page }) => {
    await navigateTo(page, path);
    // Cards deal in with a staggered slide-up animation; measure their resting positions, not
    // the last cards still sliding into place.
    // Infinite animations (e.g. a pulsing highlight) never finish, so only finite ones are awaited.
    await page.waitForFunction(() =>
      document
        .getAnimations()
        .every(
          (a) => a.playState !== 'running' || a.effect?.getComputedTiming().iterations === Number.POSITIVE_INFINITY,
        ),
    );
    const hidden = await page.evaluate((visiblePx) => {
      const footer = document.querySelector('[data-testid="game-footer-content"]');
      if (!footer) return [];
      const bottom = footer.getBoundingClientRect().bottom;
      return [...footer.querySelectorAll('img, svg[role="img"]')]
        .map((el) => el.getBoundingClientRect())
        .filter((box) => box.width >= 20 && box.height >= 30 && box.top + visiblePx > bottom)
        .map((box) => Math.round(box.top));
    }, VISIBLE_PX);
    expect(
      hidden,
      `${path}: cards whose top ${VISIBLE_PX}px is below the footer content (tops: ${hidden.join(', ')})`,
    ).toEqual([]);
  });
}
