import { expect, type Page, test } from '@playwright/test';
import { AUXILIARY_ACTION_LABELS } from '../auxiliaryActions';

/** Height of a card's top strip that must stay visible: rank corner plus a 44px tap target. */
const VISIBLE_PX = 52;
/** Controls a player can activate: buttons, button roles, and focusable custom widgets. */
const GAME_ACTION_SELECTOR = 'button, [role="button"], [tabindex="0"]';

/** Checks the first actionable game control in an already loaded page. */
export async function checkFirstAction(
  page: Page,
  path: string,
  _viewport: { width: number; height: number },
): Promise<void> {
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

  // A missing <main> fails this check softly instead of throwing, so later checks still run.
  if (!result.mainFound) {
    expect.soft(result.mainFound, `First-action check: ${path}: game main content was not found`).toBe(true);
    return;
  }
  if (path === '/hearts') {
    expect
      .soft(result.actionableCount, `First-action check: /hearts actionable game controls on initial load`)
      .toBeGreaterThan(0);
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
