import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * #11700 keeps each human card visible in the 1280×800 desktop play area. This mirrors
 * hand-visibility.spec.ts (#11655), measuring the desktop play area and its scroll ancestors.
 */
test.use({ viewport: { width: 1280, height: 800 } });
const PAGES = ['/bigtwo', '/tienlen', '/zheng', '/president', '/rook', '/bidwhist'];
const VISIBLE_PX = 52;

for (const path of PAGES) {
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
    const result = await page.evaluate((visiblePx) => {
      const cards = [
        ...document.querySelectorAll('[data-testid^="hand-card-"] img, [data-testid^="hand-card-"] [role="img"]'),
      ]
        .map((el) => el.getBoundingClientRect())
        .filter((box) => box.width >= 20 && box.height >= 30);
      const hidden = cards
        .filter((box) => {
          let visibleBottom = window.innerHeight;
          let ancestor = (document.querySelector('[data-testid^="hand-card-"]') as HTMLElement | null)?.parentElement;
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
      return { cards: cards.length, hidden };
    }, VISIBLE_PX);
    expect(result.cards, `${path}: no hand cards found`).toBeGreaterThan(0);
    expect(result.hidden, `${path}: hidden hand cards at tops ${result.hidden.join(', ')}`).toEqual([]);
  });
}
