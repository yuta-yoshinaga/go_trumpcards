import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * Representative play-area scroll budgets at 375×667. The original overflow
 * and worst post-fix measurements are separated by each limit: nertz 1125/415/770,
 * crescent 849/459/654, loba 795/486/640, sthelena 781/440/610, mendikot
 * 725/545/635, pan 718/498/608, popejoan 713/301/507, minibridge 667/319/493,
 * trex 649/237/443, marriage 623/477/550 (before/after/limit, px). The fixes
 * shipped in PRs #11622, #11623, #11624, and #11626. These heights depend on the
 * shuffled deal (#4373), so one load per page is intentional; the limits allow
 * deal variation while catching a regression to the pre-fix layout.
 */
test.use({ viewport: { width: 375, height: 667 } });

const BUDGETS: ReadonlyArray<{ path: string; maxOverflow: number }> = [
  { path: '/nertz', maxOverflow: 770 },
  { path: '/crescent', maxOverflow: 654 },
  { path: '/loba', maxOverflow: 640 },
  { path: '/sthelena', maxOverflow: 610 },
  { path: '/mendikot', maxOverflow: 635 },
  { path: '/pan', maxOverflow: 608 },
  { path: '/popejoan', maxOverflow: 507 },
  { path: '/minibridge', maxOverflow: 493 },
  { path: '/trex', maxOverflow: 443 },
  { path: '/marriage', maxOverflow: 550 },
];

for (const { path, maxOverflow } of BUDGETS) {
  test(`${path} keeps play-area scrolling within its mobile budget`, async ({ page }) => {
    await navigateTo(page, path);
    const region = page.locator('main [aria-busy] > div.flex-1.overflow-y-auto');
    const { overflow, scrollHeight, clientHeight } = await region.evaluate((element) => ({
      overflow: element.scrollHeight - element.clientHeight,
      scrollHeight: element.scrollHeight,
      clientHeight: element.clientHeight,
    }));
    expect(
      overflow,
      `${path} play area overflow is ${overflow}px (scrollHeight ${scrollHeight}px - clientHeight ${clientHeight}px; limit ${maxOverflow}px)`,
    ).toBeLessThanOrEqual(maxOverflow);
  });
}
