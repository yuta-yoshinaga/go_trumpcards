import { expect, test } from '@playwright/test';
import {
  DESKTOP_VIEWPORT_COMPACT,
  DESKTOP_VIEWPORT_LARGE,
  DESKTOP_VIEWPORTS,
  getDesktopCardWidthLimit,
  PHONE_VIEWPORT,
  VISIBLE_PX_Y,
} from './desktopViewports';
import { navigateTo, waitForAnimations, waitForLoaded } from './helpers';
import { measureBoardCards, PAGES } from './sweeps/board';

/** #11702: use the same 52px visible-card rule as #11655 for desktop boards. */

test('uses the large-desktop card widths for each desktop breakpoint', () => {
  expect(getDesktopCardWidthLimit(1024)).toBe(120);
  expect(getDesktopCardWidthLimit(1279)).toBe(120);
  expect(getDesktopCardWidthLimit(1280)).toBe(150);
});

test('defines the phone and desktop viewport breakpoints', () => {
  expect(PHONE_VIEWPORT).toEqual({ width: 375, height: 667 });
  expect(DESKTOP_VIEWPORT_LARGE).toEqual({ width: 1280, height: 800 });
  expect(DESKTOP_VIEWPORT_COMPACT).toEqual({ width: 1024, height: 800 });
  expect(DESKTOP_VIEWPORTS).toEqual([DESKTOP_VIEWPORT_LARGE, DESKTOP_VIEWPORT_COMPACT]);
});

// The fixed list also proves the card selector still finds cards (count > 0). Every other game is
// held to the same rule in route-sweep.spec.ts (#11727).
for (const viewport of DESKTOP_VIEWPORTS) {
  test.describe(`${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });
    for (const path of PAGES) {
      test(`${path} keeps every board card visible on desktop`, async ({ page }) => {
        await navigateTo(page, path);
        const result = await measureBoardCards(page, path, viewport.width);
        expect(result.count, `${path}: no board cards found`).toBeGreaterThan(0);
        expect(result.hidden, `${path}: board cards hidden below their play area`).toEqual([]);
        expect(result.hiddenX, `${path}: board cards hidden beside their play area`).toEqual([]);
        expect(result.oversized, `${path}: board cards wider than allowed (width and alt)`).toEqual([]);
      });
    }

    test('/agnes keeps every large desktop move button visible', async ({ page }) => {
      await page.setViewportSize(viewport);
      // The bar lists only the moves that are legal now, so a deal with none renders it empty.
      // Re-deal until one has a move, so the hit test below always has buttons to check.
      const moveBar = page.getByTestId('agnes-move-bar');
      await navigateTo(page, '/agnes');
      for (let deal = 0; deal < 10; deal++) {
        if (deal > 0) {
          // Each mount resets the game, so a reload is a fresh deal.
          await page.reload();
          await waitForLoaded(page);
        }
        await expect(moveBar).toBeAttached();
        if ((await moveBar.getByRole('button').count()) > 0) break;
      }
      await expect(moveBar.getByRole('button').first()).toBeVisible();

      const result = await page.evaluate(() => {
        const controls = [...document.querySelectorAll('[data-testid="agnes-move-bar"] button')];
        return controls.map((element) => {
          const box = element.getBoundingClientRect();
          const hit = document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2);
          return hit === element || element.contains(hit);
        });
      });
      expect(result.length).toBeGreaterThan(0);
      expect(result.every(Boolean)).toBe(true);
    });

    for (const { path, selector } of [
      { path: '/shelem', selector: '[data-testid="sh-primary-actions"] button' },
      { path: '/put', selector: '[data-testid="put-primary-actions"] button' },
      { path: '/truco', selector: '[data-testid="truco-primary-actions"] button' },
    ]) {
      test(`${path} keeps primary controls inside the desktop play viewport`, async ({ page }) => {
        await page.setViewportSize(viewport);
        await navigateTo(page, path);
        const controls = page.locator(selector);
        await expect(controls.first()).toBeVisible();
        const visible = await controls.evaluateAll((elements) =>
          elements.every((element) => {
            const box = element.getBoundingClientRect();
            let top = 0;
            let bottom = window.innerHeight;
            for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
              const style = getComputedStyle(ancestor);
              if (
                /^(auto|scroll|hidden|clip)$/.test(style.overflowY) &&
                ancestor.scrollHeight > ancestor.clientHeight
              ) {
                const bounds = ancestor.getBoundingClientRect();
                top = Math.max(top, bounds.top);
                bottom = Math.min(bottom, bounds.bottom);
              }
            }
            return box.top >= top && box.bottom <= bottom;
          }),
        );
        expect(visible).toBe(true);
      });
    }

    test('/clocksolitaire keeps every clock label and progress badge on top', async ({ page }) => {
      await page.setViewportSize(viewport);
      await navigateTo(page, '/clocksolitaire');
      const result = await page.evaluate(() => {
        const controls = [
          ...document.querySelectorAll('[data-testid^="clock-pile-label-"], [data-testid^="clock-pile-progress-"]'),
        ];
        return controls.map((element) => {
          const box = element.getBoundingClientRect();
          const hit = document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2);
          return hit === element || element.contains(hit);
        });
      });
      expect(result).toHaveLength(26);
      expect(result.every(Boolean)).toBe(true);
    });

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
            (images, visiblePxY) =>
              images.filter((image) => {
                const box = image.getBoundingClientRect();
                let visibleBottom = window.innerHeight;
                for (let ancestor = image.parentElement; ancestor; ancestor = ancestor.parentElement) {
                  const style = getComputedStyle(ancestor);
                  if (
                    /^(auto|scroll|hidden|clip)$/.test(style.overflowY) &&
                    ancestor.scrollHeight > ancestor.clientHeight
                  ) {
                    visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
                  }
                }
                return box.width >= 20 && box.height >= 30 && box.top + visiblePxY <= visibleBottom;
              }).length,
            VISIBLE_PX_Y,
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

    test('sixcardgolf keeps the draw and discard cards visible after a draw and placement', async ({ page }) => {
      await navigateTo(page, '/sixcardgolf');
      const hand = page.getByTestId('scg-seat-0').locator('button');
      await hand.nth(0).click();
      await hand.nth(1).click();
      await page.getByTestId('scg-draw-stock').click();
      await page.getByTestId('scg-seat-0').locator('button').first().click();
      await waitForAnimations(page);
      const stockAndDiscard = await page.evaluate(
        (visiblePxY) =>
          ['scg-stock', 'scg-discard-draw-option'].map((testId) => {
            const element = document.querySelector(`[data-tutorial="${testId}"], [data-testid="${testId}"]`);
            if (!element) return false;
            const box = element.getBoundingClientRect();
            return box.top + visiblePxY <= window.innerHeight && box.bottom > 0;
          }),
        VISIBLE_PX_Y,
      );
      expect(stockAndDiscard).toEqual([true, true]);
    });
  });
}
