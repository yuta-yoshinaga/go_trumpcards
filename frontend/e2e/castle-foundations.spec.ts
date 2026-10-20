import { expect, test } from '@playwright/test';
import { DESKTOP_VIEWPORTS } from './desktopViewports';
import { navigateTo } from './helpers';

const castlePages = [
  { path: '/fortress', foundation: '[data-tutorial="fortress-foundation"]' },
  { path: '/streetsandalleys', foundation: '[data-tutorial="sa-foundation"]' },
  { path: '/beleagueredcastle', foundation: '[data-tutorial="bc-foundation"]' },
  { path: '/citadel', foundation: '[data-tutorial="bc-foundation"]' },
] as const;

for (const viewport of DESKTOP_VIEWPORTS) {
  test.describe(`desktop ${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });

    for (const { path, foundation } of castlePages) {
      test(`${path} foundations fit`, async ({ page }) => {
        await navigateTo(page, path);
        const container = page.locator(foundation);
        await expect(container).toBeVisible();

        const measurements = await container.locator('button').evaluateAll((buttons) =>
          buttons.map((button) => {
            const rect = button.getBoundingClientRect();
            let bottom = window.innerHeight;
            for (let ancestor = button.parentElement; ancestor; ancestor = ancestor.parentElement) {
              const overflowY = getComputedStyle(ancestor).overflowY;
              if (/^(auto|scroll)$/.test(overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
                bottom = ancestor.getBoundingClientRect().bottom;
                break;
              }
            }
            return { buttonBottom: rect.bottom, boundaryBottom: bottom };
          }),
        );

        for (const [slotIndex, { buttonBottom, boundaryBottom }] of measurements.entries()) {
          expect
            .soft(
              buttonBottom,
              `${path} at ${viewport.width}x${viewport.height}: foundation slot ${slotIndex} bottom ${buttonBottom} <= boundary ${boundaryBottom}`,
            )
            .toBeLessThanOrEqual(boundaryBottom);
        }
      });
    }
  });
}
