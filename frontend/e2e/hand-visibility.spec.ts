import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { AUXILIARY_ACTION_LABELS } from './auxiliaryActions';
import { navigateTo } from './helpers';

/**
 * Every hand card stays reachable without scrolling the footer at 375×667 (#11655) and 1280×800 (#11701).
 * The phone test also checks first-action visibility in this same page load to reduce E2E shard time (#11889).
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
 * Routes come from gameRoutes.ts (#11821). The original fixed list remains a strict floor for
 * pages known to put hands in the footer, so losing their footer or cards still fails. Other
 * routes are measured when they have footer hands and annotated as skipped when they do not;
 * this two-tier guard covers newly added footer hands without treating play-area hands as defects.
 */
/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;
/** Controls a player can activate: buttons, button roles, and focusable custom widgets. */
const GAME_ACTION_SELECTOR = 'button, [role="button"], [tabindex="0"]';

// These routes have a confirmed first-action layout defect tracked by #11883.
// Their first-action assertion is inverted so a fix makes the test fail until this list is updated.
const KNOWN_FIRST_ACTION_FAILURES = ['/briscola', '/estimation', '/pasur', '/polignac', '/quodlibet', '/schnapsen'];

// These routes fail depending on the deal; omit only their first-action assertion (#11883).
const DEAL_DEPENDENT_FIRST_ACTION_FAILURES = ['/tienlen', '/minibridge', '/nainjaune', '/baloot', '/escoba'];

/**
 * Pages whose footer content overflowed at the start of #11655 (and the very large hands), minus
 * the ones whose hand lives in the play area (Crazy Four Poker, Double Attack, Double Exposure,
 * Free Bet, Michigan, Spanish 21, Trente et Quarante). These routes must keep their footer hand.
 */
const FOOTER_HAND_PAGES: readonly string[] = [
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

const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

test('route list is complete', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(300);
});

test('every strict footer-hand page is a game route', () => {
  const missingPages = FOOTER_HAND_PAGES.filter((path) => !routes.includes(path));
  expect(missingPages, `footer-hand pages missing from game routes: ${missingPages.join(', ')}`).toEqual([]);
});

test('known first-action failure routes are game routes', () => {
  for (const path of [...KNOWN_FIRST_ACTION_FAILURES, ...DEAL_DEPENDENT_FIRST_ACTION_FAILURES]) {
    expect(routes, `${path} remains a game route`).toContain(path);
  }
});

/** Enforces the 52px visibility rule, with strict presence floors for known footer-hand pages. */
async function checkHandVisibility(page: import('@playwright/test').Page, path: string) {
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

/** Checks the first actionable game control in an already loaded page. */
async function checkFirstActionVisibility(page: import('@playwright/test').Page, path: string) {
  const result = await page.evaluate(
    ({ visiblePx, auxiliaryLabels, actionSelector }) => {
      // Keep app-bar controls out of this check; only controls in the game main count.
      // This also includes GameFooter actions inside the main content.
      const main = document.querySelector('main');
      if (!main) return { mainFound: false, actionableCount: 0, visible: [], mostVisible: null };
      const controls = [...main.querySelectorAll<HTMLElement>(actionSelector)].filter((element) => {
        if (element.matches(':disabled') || element.getAttribute('aria-disabled') === 'true') return false;
        if (element.closest('[data-app-chrome], header, nav, [data-testid="phase-indicator"]')) return false;
        const label = (element.getAttribute('aria-label') || element.textContent || '').trim().replace(/\s+/g, ' ');
        if (auxiliaryLabels.includes(label)) return false;
        return true;
      });
      const isRendered = (element: HTMLElement) => {
        if (element.closest('[aria-hidden="true"], .sr-only')) return false;
        const details = element.closest('details:not([open])');
        if (details && !element.closest('summary')) return false;
        for (let ancestor: HTMLElement | null = element; ancestor; ancestor = ancestor.parentElement) {
          const style = getComputedStyle(ancestor);
          if (style.display === 'none' || style.visibility === 'hidden' || Number.parseFloat(style.opacity) < 0.05) {
            return false;
          }
        }
        const rect = element.getBoundingClientRect();
        return rect.width >= 4 && rect.height >= 4;
      };
      const visibleHeight = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        let top = Math.max(rect.top, 0);
        let bottom = Math.min(rect.bottom, window.innerHeight);
        let clipBottom: number | null = null;
        for (
          let ancestor = element.parentElement;
          ancestor && ancestor !== document.body;
          ancestor = ancestor.parentElement
        ) {
          if (/^(auto|scroll|hidden)$/.test(getComputedStyle(ancestor).overflowY)) {
            const clip = ancestor.getBoundingClientRect();
            top = Math.max(top, clip.top);
            if (clip.bottom < bottom) clipBottom = clip.bottom;
            bottom = Math.min(bottom, clip.bottom);
          }
        }
        return { height: Math.max(0, bottom - top), clipBottom };
      };
      const actionable = controls.filter(isRendered);
      const measured = actionable.map((element) => {
        const rect = element.getBoundingClientRect();
        const visibility = visibleHeight(element);
        return {
          label: (element.getAttribute('aria-label') || element.textContent || '(unlabelled)')
            .trim()
            .replace(/\s+/g, ' '),
          top: Math.round(rect.top),
          visibleHeight: visibility.height,
          height: rect.height,
          clipBottom: visibility.clipBottom === null ? null : Math.round(visibility.clipBottom),
        };
      });
      const visible = measured.filter((control) => control.visibleHeight >= Math.min(visiblePx, control.height));
      const mostVisible = [...measured].sort((a, b) => b.visibleHeight - a.visibleHeight || a.top - b.top)[0] ?? null;
      return { mainFound: true, actionableCount: actionable.length, visible, mostVisible };
    },
    { visiblePx: VISIBLE_PX, auxiliaryLabels: AUXILIARY_ACTION_LABELS, actionSelector: GAME_ACTION_SELECTOR },
  );

  // A missing <main> fails this check softly instead of throwing, so the hand check still runs.
  if (!result.mainFound) {
    expect.soft(result.mainFound, `First-action check: ${path}: game main content was not found`).toBe(true);
    return;
  }
  if (path === '/hearts') {
    expect
      .soft(result.actionableCount, `First-action check: /hearts actionable game controls on initial load`)
      .toBeGreaterThan(0);
  }
  if (DEAL_DEPENDENT_FIRST_ACTION_FAILURES.includes(path)) {
    test.info().annotations.push({
      type: 'known first-action issue',
      description: `${path}: first-action visibility depends on the deal (#11883).`,
    });
    return;
  }
  if (KNOWN_FIRST_ACTION_FAILURES.includes(path)) {
    expect
      .soft(
        result.visible,
        `First-action check: ${path}: known #11883 failure; remove this route when fixed. ${VISIBLE_PX}px visible control expected to remain absent: ${JSON.stringify(result)}`,
      )
      .toHaveLength(0);
    return;
  }
  if (result.actionableCount === 0) {
    test.info().annotations.push({
      type: 'skipped first-action check',
      description: 'No actionable game controls on initial load (for example, a CPU turn).',
    });
    return;
  }
  expect
    .soft(
      result.visible,
      `First-action check: ${path}: no actionable control has ${VISIBLE_PX}px visible (or its full height when shorter): ${JSON.stringify(result)}`,
    )
    .not.toHaveLength(0);
}

test.describe('phone 375x667', () => {
  test.use({ viewport: { width: 375, height: 667 } });
  for (const path of routes) {
    test(`${path} keeps every footer hand card reachable on a phone`, async ({ page }) => {
      await navigateTo(page, path);
      await checkFirstActionVisibility(page, path);
      await checkHandVisibility(page, path);
    });
  }
});

test.describe('desktop 1280x800', () => {
  test.use({ viewport: { width: 1280, height: 800 } });
  for (const path of routes) {
    test(`${path} keeps every footer hand card reachable on a desktop`, async ({ page }) => {
      await navigateTo(page, path);
      await checkHandVisibility(page, path);
    });
  }
});
