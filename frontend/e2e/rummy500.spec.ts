import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Rummy 500 E2E', () => {
  test('navigates and renders the basic UI', async ({ page }) => {
    await navigateTo(page, '/rummy500');
    await expect(page.getByText(/ラウンド\s*1/)).toBeVisible();
    await expect(page.getByText(/山札:/)).toBeVisible();
  });

  test('shows the draw stock button in Draw phase', async ({ page }) => {
    await navigateTo(page, '/rummy500');
    const drawBtn = page.getByRole('button', { name: '山札から引く' });
    await expect(drawBtn).toBeVisible();
  });

  test('clicking draw stock advances to Play phase and shows action buttons', async ({ page }) => {
    await navigateTo(page, '/rummy500');
    await page.getByRole('button', { name: '山札から引く' }).click();
    await waitForLoaded(page);
    // The "discard" or "meld" controls appear in Play phase.
    const meldBtn = page.getByRole('button', { name: 'メルドする' });
    await expect(meldBtn).toBeVisible();
  });

  test('reset restarts the game', async ({ page }) => {
    await navigateTo(page, '/rummy500');
    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await resetButton.click();
    // Confirm dialog appears
    const confirmButton = page.getByRole('button', { name: '確認' });
    if (await confirmButton.isVisible()) {
      await confirmButton.click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/ラウンド\s*1/)).toBeVisible();
  });
});
