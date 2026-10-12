import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * #11700 keeps each human card visible in the 1280×800 desktop play area. This mirrors
 * hand-visibility.spec.ts (#11655), measuring the desktop play area and its scroll ancestors.
 */
test.use({ viewport: { width: 1280, height: 800 } });
const PAGES = [
  { path: '/bigtwo', hand: '[data-tutorial="bt-player-hand"]' },
  { path: '/tienlen', hand: '[data-tutorial="tl-player-hand"]' },
  { path: '/zheng', hand: '[data-tutorial="zheng-player-hand"]' },
  { path: '/president', hand: '[data-tutorial="pr-player-hand"]' },
  { path: '/rook', hand: '[data-tutorial="rook-hand"]' },
  { path: '/bidwhist', hand: '[data-tutorial="bw-hand"]' },
  { path: '/mendikot', hand: '[data-tutorial="md-hand"]' },
  { path: '/honeymoonbridge', hand: '[data-tutorial="hb-hand"]' },
  { path: '/reversis', hand: '[data-tutorial="reversis-hand"]' },
  { path: '/goofspiel', hand: '[data-tutorial="gs-hand"]' },
  { path: '/scopone', hand: '[data-tutorial="sp-player-hand"]' },
  { path: '/kalooki', hand: '[data-tutorial="kalooki-hand"]' },
  { path: '/contractrummy', hand: '[data-testid="cr-hand"]' },
  { path: '/nainjaune', hand: '[data-tutorial="nj-hand"]' },
  { path: '/sjavs', hand: '[data-tutorial="sj-hand"]' },
  { path: '/mushi', hand: '[data-tutorial="mushi-hand"]' },
  { path: '/desmoche', hand: '[data-tutorial="ds-hand"]' },
  { path: '/fiftyone', hand: '[data-tutorial="fo-player-hand"]' },
  { path: '/tusac', hand: '[data-tutorial="tusac-hand"]' },
  { path: '/pitch', hand: '[data-tutorial="pt-player-hand"]' },
  { path: '/ironcross', hand: '[data-tutorial="ic-hand"]' },
  { path: '/loba', hand: '[data-tutorial="lb-hand"]' },
  { path: '/spiteandmalice', hand: '[data-tutorial="sam-hand"]' },
  { path: '/chineseten', hand: '[data-tutorial="ct-hand"]' },
  { path: '/carioca', hand: '[data-tutorial="ca-hand"]' },
  { path: '/poch', hand: '[data-tutorial="pc-hand"]' },
  { path: '/bura', hand: '[data-tutorial="bura-hand"]' },
  { path: '/karnoffel', hand: '[data-tutorial="karnoffel-hand"]' },
  { path: '/binokel', hand: '[data-tutorial="bn-player-hand"]' },
];
const VISIBLE_PX = 52;

for (const { path, hand } of PAGES) {
  test(`${path} keeps desktop hand cards visible`, async ({ page }) => {
    await navigateTo(page, path);
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
    const result = await page.evaluate(
      ({ visiblePx, handSelector }) => {
        const handContainer = document.querySelector(handSelector);
        if (!handContainer) return { missingHand: true as const, cards: 0, hidden: [] as number[] };
        const cards = [...handContainer.querySelectorAll('img, [role="img"]')]
          .map((el) => el.getBoundingClientRect())
          .filter((box) => box.width >= 20 && box.height >= 30);
        const hidden = cards
          .filter((box) => {
            let visibleBottom = window.innerHeight;
            let ancestor = handContainer.parentElement;
            while (ancestor) {
              const style = getComputedStyle(ancestor);
              if (
                ['auto', 'scroll', 'hidden'].includes(style.overflowY) &&
                ancestor.scrollHeight > ancestor.clientHeight
              ) {
                visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
              }
              ancestor = ancestor.parentElement;
            }
            return box.top + visiblePx > visibleBottom;
          })
          .map((box) => Math.round(box.top));
        return { missingHand: false as const, cards: cards.length, hidden };
      },
      { visiblePx: VISIBLE_PX, handSelector: hand },
    );
    expect(result.missingHand, `${path}: hand container ${hand} not found`).toBe(false);
    expect(result.cards, `${path}: no hand cards found`).toBeGreaterThan(0);
    expect(result.hidden, `${path}: hidden hand cards at tops ${result.hidden.join(', ')}`).toEqual([]);
  });
}
