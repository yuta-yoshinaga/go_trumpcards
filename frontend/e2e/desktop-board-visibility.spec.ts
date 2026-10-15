import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, type Page, test } from '@playwright/test';
import { DESKTOP_VIEWPORTS, KNOWN_1024_FAILURES, KNOWN_HORIZONTAL_FAILURES, VISIBLE_PX_X } from './desktopViewports';
import { navigateTo, waitForAnimations, waitForLoaded } from './helpers';

/** #11702: use the same 52px visible-card rule as #11655 for desktop boards. */

const PAGES = [
  '/grandfathersclock',
  '/agnes',
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
  '/openfacechinese',
  '/canasta',
] as const;

// Read the route declarations as text so this spec does not load the app's React modules.
const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const allGamePaths = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

// A tight floor (383 routes today): a regex or file-layout change that drops games fails here by name
// instead of silently shrinking the sweep.
test('reads the full game route list for the all-games sweep', () => {
  expect(allGamePaths.length, 'game route path declarations found').toBeGreaterThanOrEqual(360);
});

// #11700 excludes CPU-seat cards, face-up cards inside accordions, stud door cards, and discard history.
// Samba is excluded because opposing-team meld cards can fall below the play area depending on the deal.
const excludedGamePaths = new Set([
  '/soko',
  '/fivecardstud',
  '/sevencardstud',
  '/sevencardstudhilo',
  '/chicago',
  '/nertz',
  '/pan',
  '/bolivia',
  '/indianrummy',
  '/burraco',
  '/biriba',
  '/baseballpoker',
  '/samba',
  // The last discarded pair is discard history, which #11700 left out of the rule.
  '/oldmaid',
]);
// Accordion's 52-card row is an intentional horizontal scroller, so only the horizontal rule skips it.
const horizontalScrollExemptions = new Set(['/accordion']);
/**
 * Waits for the deal animation, then returns the face-up board cards, those whose top 52px is hidden,
 * and those with less than VISIBLE_PX_X showing beside a horizontally clipping ancestor.
 */
async function measureBoardCards(page: Page, path: string, width: number) {
  await waitForAnimations(page);
  const skipHorizontal =
    horizontalScrollExemptions.has(path) || (KNOWN_HORIZONTAL_FAILURES.get(path)?.has(width) ?? false);
  return page.evaluate(
    ({ visiblePxX, skipHorizontal }) => {
      const cards = [...document.querySelectorAll('main img, main [role="img"]')].filter((element) => {
        const label = `${element.getAttribute('alt') ?? ''} ${element.getAttribute('aria-label') ?? ''}`;
        const box = element.getBoundingClientRect();
        // CPU hands inside the shared accordion are reference only (#11700); they may scroll.
        if (element.closest('[data-testid="cpu-accordion"]')) return false;
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
      const hiddenX = skipHorizontal
        ? []
        : cards.flatMap((element) => {
            const box = element.getBoundingClientRect();
            let visibleLeft = 0;
            let visibleRight = window.innerWidth;
            for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
              const style = getComputedStyle(ancestor);
              if (
                /^(auto|scroll|hidden|clip)$/.test(style.overflowX) &&
                ancestor.scrollWidth > ancestor.clientWidth + 1
              ) {
                const bounds = ancestor.getBoundingClientRect();
                visibleLeft = Math.max(visibleLeft, bounds.left);
                visibleRight = Math.min(visibleRight, bounds.right);
              }
            }
            return box.left + visiblePxX > visibleRight || box.right - visiblePxX < visibleLeft
              ? [Math.round(box.left)]
              : [];
          });
      return { count: cards.length, hidden, hiddenX };
    },
    { visiblePxX: VISIBLE_PX_X, skipHorizontal },
  );
}

// The fixed list also proves the card selector still finds cards (count > 0). Every other game is
// held to the same rule (#11727); a game may open with no face-up card, so only "none hidden" applies there.
for (const viewport of DESKTOP_VIEWPORTS) {
  test.describe(`${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });
    const skipKnownFailure = (path: string) => {
      if (viewport.width === 1024 && KNOWN_1024_FAILURES.board.has(path))
        test.skip(true, '#11811: known 1024px visibility failure');
    };
    for (const path of PAGES) {
      test(`${path} keeps every board card visible on desktop`, async ({ page }) => {
        skipKnownFailure(path);
        await navigateTo(page, path);
        const result = await measureBoardCards(page, path, viewport.width);
        expect(result.count, `${path}: no board cards found`).toBeGreaterThan(0);
        expect(result.hidden, `${path}: board cards hidden below their play area`).toEqual([]);
        expect(result.hiddenX, `${path}: board cards hidden beside their play area`).toEqual([]);
      });
    }

    const fixedPaths = new Set<string>(PAGES);
    for (const path of allGamePaths.filter((p) => !fixedPaths.has(p) && !excludedGamePaths.has(p))) {
      test(`${path} (all-games sweep) keeps every board card visible on desktop`, async ({ page }) => {
        skipKnownFailure(path);
        await navigateTo(page, path);
        const result = await measureBoardCards(page, path, viewport.width);
        expect(result.hidden, `${path}: board cards hidden below their play area`).toEqual([]);
        expect(result.hiddenX, `${path}: board cards hidden beside their play area`).toEqual([]);
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
        skipKnownFailure(path);
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
              if (/^(auto|scroll|hidden)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
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
        skipKnownFailure(path);
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
  });
}
