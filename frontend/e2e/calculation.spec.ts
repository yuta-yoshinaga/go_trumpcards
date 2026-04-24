import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Calculation E2E', () => {
  test('navigates to calculation and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/calculation');

    // Wait for the move-count readout
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Foundations row: 4 foundations with step labels
    await expect(page.getByLabel(/Foundation 0 \+1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByLabel(/Foundation 3 \+4/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Control buttons while playing
    await expect(page.getByRole('button', { name: 'ヒント' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button targeted via data-tutorial
    await expect(page.locator('[data-tutorial="ca-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/calculation');
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();

    // After giveup, the action log button should appear
    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
