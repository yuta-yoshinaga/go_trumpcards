import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * Representative play-area scroll budgets at 375×667 (#11613). Each limit is the
 * midpoint between the page's overflow before its fix (PRs #11622, #11623, #11624,
 * #11626, then #11636, #11637, #11638, then #11642, #11643, #11644, #11645) and the
 * worst overflow measured after it; the measurements are in #11627 (first ten pages),
 * #11640 (the next eight) and this PR's description (the round-three pages).
 * These heights depend on the shuffled deal (#4373), so one load per page is
 * intentional and the limits leave room for deal variation.
 *
 * The guard is statistical for /loba, /mendikot, /minibridge and /kille: their height varies
 * so much by deal that the pre-fix layout passed on some deals. A pass there is not a
 * proof, and a failure means "re-measure the worst case", never "retry".
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
  { path: '/desmoche', maxOverflow: 439 },
  { path: '/karnoffel', maxOverflow: 416 },
  { path: '/jass', maxOverflow: 407 },
  { path: '/crazyquilt', maxOverflow: 375 },
  { path: '/openfacechinese', maxOverflow: 373 },
  { path: '/slyfox', maxOverflow: 332 },
  { path: '/nainjaune', maxOverflow: 314 },
  { path: '/boston', maxOverflow: 307 },
  { path: '/kille', maxOverflow: 422 },
  { path: '/carioca', maxOverflow: 359 },
  { path: '/julepe', maxOverflow: 329 },
  { path: '/colorado', maxOverflow: 329 },
  { path: '/stealingbundles', maxOverflow: 314 },
  { path: '/oldmaid', maxOverflow: 302 },
  { path: '/indianrummy', maxOverflow: 301 },
  { path: '/tarabish', maxOverflow: 300 },
  { path: '/tute', maxOverflow: 281 },
  { path: '/rollingstone', maxOverflow: 268 },
  { path: '/klaverjas', maxOverflow: 234 },
  { path: '/poch', maxOverflow: 230 },
  { path: '/yaniv', maxOverflow: 228 },
  { path: '/pasur', maxOverflow: 228 },
  { path: '/belote', maxOverflow: 220 },
  { path: '/goofspiel', maxOverflow: 213 },
  { path: '/vint', maxOverflow: 208 },
  { path: '/crazyeights', maxOverflow: 205 },
  { path: '/hasenpfeffer', maxOverflow: 184 },
  { path: '/laughandliedown', maxOverflow: 173 },
  { path: '/cucumber', maxOverflow: 157 },
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
