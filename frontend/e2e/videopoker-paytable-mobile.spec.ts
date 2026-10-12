import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Video Poker paytable on mobile', () => {
  test('fits the payout table within a 375px viewport for all variants', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });

    for (const game of ['videopoker', 'deuceswild', 'jokerpoker']) {
      await navigateTo(page, `/${game}`);
      await waitForLoaded(page);

      const table = page.locator('table[aria-label]').first();
      await expect(table).toBeVisible();
      expect(await table.evaluate((element) => element.getBoundingClientRect().right)).toBeLessThanOrEqual(375);
    }
  });
});
