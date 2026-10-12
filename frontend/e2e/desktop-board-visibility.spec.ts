import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

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
] as const;

for (const path of PAGES) {
  test(`${path} keeps every board card visible on desktop`, async ({ page }) => {
    await navigateTo(page, path);
    await page.waitForFunction(() =>
      document
        .getAnimations()
        .every(
          (animation) =>
            animation.playState === 'finished' ||
            animation.playState === 'idle' ||
            animation.effect?.getComputedTiming().iterations === Number.POSITIVE_INFINITY,
        ),
    );
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

const POKER_BOARDS = [
  { path: '/omaha', selector: '[data-tutorial="oh-community-cards"]', initial: 0 },
  { path: '/dramaha', selector: '[data-tutorial="dr-community-cards"]', initial: 0 },
  // Courchevel exposes one community card pre-flop.
  { path: '/courchevel', selector: '[data-tutorial="bo-community-cards"]', firstCardShownPreflop: 1 },
] as const;

for (const { path, selector, firstCardShownPreflop = 0 } of POKER_BOARDS) {
  test(`${path} keeps the flop and river board visible on desktop`, async ({ page }) => {
    await navigateTo(page, path);
    const board = page.locator(selector);
    const visibleBoardCount = async () =>
      board.locator('img').evaluateAll(
        (images) =>
          images.filter((image) => {
            const box = image.getBoundingClientRect();
            let visibleBottom = window.innerHeight;
            for (let ancestor = image.parentElement; ancestor; ancestor = ancestor.parentElement) {
              const style = getComputedStyle(ancestor);
              if (/^(auto|scroll|hidden)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
                visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
              }
            }
            return box.width >= 20 && box.height >= 30 && box.top + 52 <= visibleBottom;
          }).length,
      );

    await expect(board.locator('img')).toHaveCount(firstCardShownPreflop === 1 ? 1 : 5, { timeout: 15_000 });
    if (firstCardShownPreflop === 1) expect(await visibleBoardCount()).toBe(1);

    let sawFlop = false;
    let sawRiver = false;
    for (let step = 0; step < 20; step++) {
      const count = await board.locator('img').count();
      if (count >= 3) {
        sawFlop = true;
        expect(await visibleBoardCount(), `${path}: flop cards must be visible`).toBe(count);
      }
      if (count === 5) {
        sawRiver = true;
        expect(await visibleBoardCount(), `${path}: river cards must be visible`).toBe(5);
        break;
      }
      const action = page.locator('button[aria-keyshortcuts="c"], button[aria-keyshortcuts="k"]').first();
      await expect(action).toBeVisible({ timeout: 30_000 });
      await action.click();
      await waitForLoaded(page);
    }
    expect(sawFlop, `${path}: deterministic check/call actions should reach the flop`).toBe(true);
    expect(sawRiver, `${path}: deterministic check/call actions should reach the river`).toBe(true);
  });
}
