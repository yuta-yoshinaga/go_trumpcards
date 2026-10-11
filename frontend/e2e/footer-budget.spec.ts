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
  // Second wave (#11668): ShengJi is a partial fix (191 → 147), so its limit is the midpoint.
  { path: '/crazyfourpoker', maxOverflow: 200 },
  { path: '/shengji', maxOverflow: 169 },
  { path: '/poker', maxOverflow: 101 },
  // Third wave (#11669) left Minchiate partial (229 → 186); stacked rows and the tall footer (#11655) bring it to 2px.
  { path: '/minchiate', maxOverflow: 115 },
  { path: '/literature', maxOverflow: 167 },
  { path: '/bauernschnapsen', maxOverflow: 129 },
  { path: '/kaiser', maxOverflow: 97 },
];

for (const { path, maxOverflow } of BUDGETS) {
  test(`${path} keeps footer-content scrolling within its mobile budget`, async ({ page }) => {
    await navigateTo(page, path);
    const region = page.getByTestId('game-footer-content');
    // navigateTo already waits for the loaded state. Do not wait for a card <img> here: Crazy Four
    // Poker has no cards before betting, and Minchiate draws its tarot cards without <img>.
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
