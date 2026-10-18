import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { AUXILIARY_ACTION_LABELS } from './auxiliaryActions';
import { navigateTo } from './helpers';

/** Minimum vertical area of the first actionable control that must be on screen. */
const VISIBLE_PX = 52;
/** Controls a player can activate: buttons, button roles, and focusable custom widgets. */
const GAME_ACTION_SELECTOR = 'button, [role="button"], [tabindex="0"]';

// These routes have a confirmed first-action layout defect tracked by #11883.
// test.fail makes a fixed route an unexpected pass, so remove it from this list when fixed.
const KNOWN_FIRST_ACTION_FAILURES = [
  '/bisley',
  '/briscola',
  '/estimation',
  '/freecell',
  '/grandfathersclock',
  '/pasur',
  '/polignac',
  '/quodlibet',
  '/schnapsen',
];

// These routes fail depending on the deal; use fixme because test.fail would reject passing deals.
const DEAL_DEPENDENT_FIRST_ACTION_FAILURES = ['/tienlen', '/minibridge', '/nainjaune', '/baloot', '/escoba'];

const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

test('route list is complete', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(360);
  for (const path of [...KNOWN_FIRST_ACTION_FAILURES, ...DEAL_DEPENDENT_FIRST_ACTION_FAILURES]) {
    expect(routes, `${path} remains a game route`).toContain(path);
  }
});

test.describe('phone 375x667', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  for (const path of routes) {
    test(`${path} shows a first game action without scrolling`, async ({ page }) => {
      if (KNOWN_FIRST_ACTION_FAILURES.includes(path)) {
        test.fail(true, 'First game action is below the fold at 375×667 (#11883)');
      }
      if (DEAL_DEPENDENT_FIRST_ACTION_FAILURES.includes(path)) {
        test.fixme(true, 'First game action visibility depends on the deal (#11883)');
      }
      await navigateTo(page, path);
      const result = await page.evaluate(
        ({ visiblePx, auxiliaryLabels, actionSelector }) => {
          // Keep app-bar controls out of this check; only controls in the game main count.
          // This also includes GameFooter actions inside the main content.
          const main = document.querySelector('main');
          if (!main) throw new Error('Game main content was not found');
          const controls = [...main.querySelectorAll<HTMLElement>(actionSelector)].filter((element) => {
            if (element.matches(':disabled') || element.getAttribute('aria-disabled') === 'true') return false;
            // Shared app chrome is identifiable by structure; footer actions remain included because
            // that region also contains game-specific controls.
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
              if (
                style.display === 'none' ||
                style.visibility === 'hidden' ||
                Number.parseFloat(style.opacity) < 0.05
              ) {
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
          const mostVisible =
            [...measured].sort((a, b) => b.visibleHeight - a.visibleHeight || a.top - b.top)[0] ?? null;
          return { actionableCount: actionable.length, visible, mostVisible };
        },
        {
          visiblePx: VISIBLE_PX,
          auxiliaryLabels: AUXILIARY_ACTION_LABELS,
          actionSelector: GAME_ACTION_SELECTOR,
        },
      );

      if (path === '/hearts') {
        expect(result.actionableCount, '/hearts actionable game controls on initial load').toBeGreaterThan(0);
      }
      if (result.actionableCount === 0) {
        test.info().annotations.push({
          type: 'skipped page',
          description: 'No actionable game controls on initial load (for example, a CPU turn).',
        });
        test.skip(true, 'No actionable game controls on initial load.');
      }
      expect(
        result.visible,
        `${path}: no actionable control has ${VISIBLE_PX}px visible (or its full height when shorter): ${JSON.stringify(result)}`,
      ).not.toHaveLength(0);
    });
  }
});
