import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Go Fish E2E', () => {
  test('navigates, resets, and verifies basic game UI', async ({ page }) => {
    await navigateTo(page, '/gofish');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game elements are visible
    // Player hand area should be present
    const playerCards = page.locator('[data-tutorial="gf-player-hand"]');
    await expect(playerCards).toBeVisible({ timeout: TIMEOUT_ACTION });

    // CPU player areas should be present
    const cpuAreas = page.locator('[data-tutorial="gf-cpu-area"]');
    await expect(cpuAreas).toBeVisible({ timeout: TIMEOUT_ACTION });

    // Ask button should be present (may be disabled)
    const askButton = page.getByRole('button', { name: '要求する' });
    await expect(askButton).toBeVisible();

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(playerCards).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
