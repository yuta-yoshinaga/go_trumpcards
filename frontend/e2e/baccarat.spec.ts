import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Baccarat E2E', () => {
  test('plays a round: bet → result → reset', async ({ page }) => {
    await navigateTo(page, '/baccarat');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // END phase: リセット button should be visible
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
