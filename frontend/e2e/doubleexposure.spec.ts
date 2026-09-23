import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Double Exposure E2E', () => {
  test('shows both dealer cards and a non-zero dealer score after betting', async ({ page }) => {
    await navigateTo(page, '/doubleexposure');

    const betButton = page.getByRole('button', { name: 'ベット', exact: true });
    await expect(betButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await betButton.click();
    await waitForLoaded(page);

    const dealerArea = page.locator('[data-tutorial="bj-dealer-hand"]');
    await expect(dealerArea).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Double Exposure reveals both dealer cards during the round.
    await expect(dealerArea.locator('img:not([src="/images/z01.png"])')).toHaveCount(2);

    const dealerScore = dealerArea.locator('p').filter({ hasText: /スコア|Score/ });
    await expect(dealerScore).toHaveText(/(?:スコア|Score)\s+[1-9][0-9]*/);
    await expect(dealerArea.locator('img[src="/images/z01.png"]')).toHaveCount(0);
  });
});
