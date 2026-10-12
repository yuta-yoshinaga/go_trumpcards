import { expect, test } from '@playwright/test';
import { navigateTo, waitForAnimations } from './helpers';

/** #11702: use the same 52px visible-card rule as #11655 for desktop boards. */
test.use({ viewport: { width: 1280, height: 800 } });

const PAGES = [
  '/grandfathersclock',
  '/bigben',
  '/colorado',
  '/slyfox',
  '/penguin',
  '/kingalbert',
  '/americantoad',
  '/terrace',
  '/crescent',
  '/sthelena',
  '/pyramid',
  '/montecarlo',
  '/eightoff',
  '/matrimony',
  '/royalcotillion',
  '/flowergarden',
  '/alaska',
  '/braid',
  '/osmosis',
  '/crazyquilt',
  '/sevens',
  '/beleagueredcastle',
  '/citadel',
  '/russiansolitaire',
  '/clocksolitaire',
  '/indianpoker',
  '/sixcardgolf',
] as const;

for (const path of PAGES) {
  test(`${path} keeps every board card visible on desktop`, async ({ page }) => {
    await navigateTo(page, path);
    await waitForAnimations(page);
    const result = await page.evaluate(() => {
      const cards = [...document.querySelectorAll('main img, main [role="img"]')].filter((element) => {
        const label = `${element.getAttribute('alt') ?? ''} ${element.getAttribute('aria-label') ?? ''}`;
        const box = element.getBoundingClientRect();
        return box.width >= 20 && box.height >= 30 && !/裏|back/i.test(label);
      });
      const hidden = cards.flatMap((element) => {
        const box = element.getBoundingClientRect();
        let visibleBottom = window.innerHeight;
        for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
          const style = getComputedStyle(ancestor);
          if (/^(auto|scroll|hidden)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
            visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
          }
        }
        return box.top + 52 > visibleBottom ? [Math.round(box.top)] : [];
      });
      return { count: cards.length, hidden };
    });
    expect(result.count, `${path}: no board cards found`).toBeGreaterThan(0);
    expect(result.hidden, `${path}: board cards hidden below their play area`).toEqual([]);
  });
}

test('sixcardgolf keeps the draw and discard cards visible after a draw and placement', async ({ page }) => {
  await navigateTo(page, '/sixcardgolf');
  const hand = page.getByTestId('scg-seat-0').locator('button');
  await hand.nth(0).click();
  await hand.nth(1).click();
  await page.getByTestId('scg-draw-stock').click();
  await page.getByTestId('scg-seat-0').locator('button').first().click();
  await waitForAnimations(page);
  const stockAndDiscard = await page.evaluate(() =>
    ['scg-stock', 'scg-discard-draw-option'].map((testId) => {
      const element = document.querySelector(`[data-tutorial="${testId}"], [data-testid="${testId}"]`);
      if (!element) return false;
      const box = element.getBoundingClientRect();
      return box.top + 52 <= window.innerHeight && box.bottom > 0;
    }),
  );
  expect(stockAndDiscard).toEqual([true, true]);
});
