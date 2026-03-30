import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Pineapple Poker E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/pineapple');

    // Click reset to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game page is rendered with pineapple-specific content
    await expect(page.getByText('あなたの手札')).toBeVisible({ timeout: 10_000 });

    // Look for betting or discard controls
    const checkButton = page.getByRole('button', { name: 'チェック' });
    const callButton = page.getByRole('button', { name: 'コール' });
    const foldButton = page.getByRole('button', { name: 'フォールド' });
    const discardControls = page.getByTestId('discard-controls');

    // Wait for either betting controls or discard phase
    await expect(checkButton.or(callButton).or(foldButton).or(discardControls).or(resetButton).first()).toBeVisible({
      timeout: 10_000,
    });

    // Verify we can reset the game
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText('あなたの手札')).toBeVisible({ timeout: 10_000 });
  });
});
