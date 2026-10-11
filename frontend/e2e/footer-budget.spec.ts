import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * Footer-content scroll budgets at 375×667 (#11655). Since #11611 pins `actions` to the
 * bottom of `GameFooter`, the rest of the footer (`game-footer-content`) is a nested
 * scroller inside the 45vh cap, and a hand placed there can be hidden even when the play
 * area does not scroll at all — play-area-budget.spec.ts cannot see it.
 *
 * Each limit is the midpoint between the page's footer overflow before its fix and the
 * worst value measured after it with `.claude/commands/scripts/play-area-overflow.sh`
 * (`mobile_footer` column); the measurements are in the PR that added the page.
 * These heights depend on the shuffled deal (#4373), so one load per page is intentional.
 */
test.use({ viewport: { width: 375, height: 667 } });

const BUDGETS: ReadonlyArray<{ path: string; maxOverflow: number }> = [
  { path: '/king', maxOverflow: 214 },
  { path: '/watten', maxOverflow: 197 },
  { path: '/cinch', maxOverflow: 187 },
  { path: '/mighty', maxOverflow: 115 },
];

for (const { path, maxOverflow } of BUDGETS) {
  test(`${path} keeps footer-content scrolling within its mobile budget`, async ({ page }) => {
    await navigateTo(page, path);
    const region = page.getByTestId('game-footer-content');
    // Measure once the hand has rendered: before that the footer is shorter than it will be.
    await expect(region.locator('img').first()).toBeVisible();
    const { overflow, scrollHeight, clientHeight } = await region.evaluate((element) => ({
      overflow: element.scrollHeight - element.clientHeight,
      scrollHeight: element.scrollHeight,
      clientHeight: element.clientHeight,
    }));
    expect(
      overflow,
      `${path} footer content overflow is ${overflow}px (scrollHeight ${scrollHeight}px - clientHeight ${clientHeight}px; limit ${maxOverflow}px)`,
    ).toBeLessThanOrEqual(maxOverflow);
  });
}
