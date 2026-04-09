import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Durak E2E', () => {
  test('starts a game: reset → verify controls → reset', async ({ page }) => {
    await navigateTo(page, '/durak');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game heading is visible (Japanese locale: ドゥラーク)
    await expect(page.getByRole('heading', { name: 'ドゥラーク' })).toBeVisible({ timeout: 10_000 });

    // Game is running, reset to start fresh
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game heading is still visible after reset
    await expect(page.getByRole('heading', { name: 'ドゥラーク' })).toBeVisible({ timeout: 10_000 });
  });
});
