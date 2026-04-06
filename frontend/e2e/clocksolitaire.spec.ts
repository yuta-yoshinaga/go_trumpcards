import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Clock Solitaire E2E', () => {
  test('navigates, resets, and plays steps', async ({ page }) => {
    await navigateTo(page, '/clocksolitaire');

    // Verify step count is displayed
    await expect(page.getByText(/ステップ数/)).toBeVisible();

    // Click step button
    const stepButton = page.getByRole('button', { name: 'ステップ' });
    await expect(stepButton).toBeVisible();
    await stepButton.click();
    await waitForLoaded(page);

    // Click step a few more times
    for (let i = 0; i < 3; i++) {
      await stepButton.click();
      await waitForLoaded(page);
    }

    // Click reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game restarted
    await expect(page.getByText(/ステップ数/)).toBeVisible();
  });

  test('autoplay completes the game', async ({ page }) => {
    await navigateTo(page, '/clocksolitaire');

    // Click autoplay
    const autoplayButton = page.getByRole('button', { name: 'オートプレイ' });
    await expect(autoplayButton).toBeVisible();
    await autoplayButton.click();
    await waitForLoaded(page);

    // After autoplay, step/autoplay buttons should not be visible
    await expect(page.getByRole('button', { name: 'ステップ' })).not.toBeVisible();
  });
});
