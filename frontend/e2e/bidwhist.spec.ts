import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Bid Whist E2E', () => {
  test('starts a game: reset → bid controls → pass → reset', async ({ page }) => {
    await navigateTo(page, '/bidwhist');

    // Reset to start a fresh game.
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // On the human's bid turn the pass button is available.
    const passButton = page.getByRole('button', { name: 'パス' });
    await expect(passButton).toBeVisible({ timeout: 10_000 });

    // Pass and let the round progress.
    await passButton.click();
    await waitForLoaded(page);

    // Reset to verify the game can be restarted.
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(passButton).toBeVisible({ timeout: 10_000 });
  });
});
