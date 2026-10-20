import { expect, type Page, test } from '@playwright/test';
import { BOARD_CARD_QUERY_SELECTOR } from '../../src/utils/boardCardSelector';
import { getDesktopCardWidthLimit, VISIBLE_PX_X, VISIBLE_PX_Y } from '../desktopViewports';
import { waitForAnimations } from '../helpers';

export const PAGES = [
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

/** Returns whether the path is in the fixed board page list, which desktop-board-visibility.spec.ts covers. */
export function isFixedBoardPage(path: string): boolean {
  return (PAGES as readonly string[]).includes(path);
}

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
 * those with less than VISIBLE_PX_X showing beside a horizontally clipping ancestor, and cards
 * whose rendered width exceeds the viewport's large-desktop card width by more than 1.5x.
 */
export async function measureBoardCards(page: Page, path: string, viewportWidth?: number) {
  let settled = true;
  try {
    await waitForAnimations(page);
  } catch {
    settled = false;
  }
  expect.soft(settled, `Board check: ${path}: animations did not settle`).toBe(true);
  const skipHorizontal = horizontalScrollExemptions.has(path);
  return page.evaluate(
    ({ visiblePxX, visiblePxY, skipHorizontal, maxCardWidth, cardSelector }) => {
      // A role=img node containing card images is a visual group (for example Salic Law's retired
      // queens), not another card. Standalone role=img cards such as tarot remain in the selector.
      const cards = [...document.querySelectorAll(cardSelector)].filter((element) => {
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
          if (/^(auto|scroll|hidden|clip)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
            visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
          }
        }
        return box.top + visiblePxY > visibleBottom ? [Math.round(box.top)] : [];
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
                // Ignore subpixel rounding noise in scroll/client width measurements.
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
      const oversized = cards.flatMap((element) => {
        const width = element.getBoundingClientRect().width;
        return width > maxCardWidth ? [{ width: Math.round(width), alt: element.getAttribute('alt') ?? '' }] : [];
      });
      return { count: cards.length, hidden, hiddenX, oversized };
    },
    {
      visiblePxX: VISIBLE_PX_X,
      visiblePxY: VISIBLE_PX_Y,
      skipHorizontal,
      maxCardWidth: getDesktopCardWidthLimit(viewportWidth ?? 1280),
      cardSelector: BOARD_CARD_QUERY_SELECTOR,
    },
  );
}

/** Enforces the 52px visible-card rule for desktop board cards. */
export async function checkBoard(page: Page, path: string, viewport: { width: number; height: number }): Promise<void> {
  if (excludedGamePaths.has(path)) {
    test.info().annotations.push({
      type: 'skipped page',
      description: `Board check: ${path}: excluded from board visibility checks (#11700).`,
    });
    return;
  }

  const result = await measureBoardCards(page, path, viewport.width);
  if (isFixedBoardPage(path)) {
    expect.soft(result.count, `Board check: ${path}: no board cards found`).toBeGreaterThan(0);
  }
  expect.soft(result.hidden, `Board check: ${path}: board cards hidden below their play area`).toEqual([]);
  expect.soft(result.hiddenX, `Board check: ${path}: board cards hidden beside their play area`).toEqual([]);
  expect.soft(result.oversized, `Board check: ${path}: board cards wider than allowed (width and alt)`).toEqual([]);
}
