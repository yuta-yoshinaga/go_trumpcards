import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('SirTommy E2E', () => {
  test('navigates to sirtommy and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/sirtommy');

    // Wait for the move-count readout
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Foundations row. Use the JA-locale label ("ファンデーション") since PR #1971
    // localised the aria-label and the Playwright suite runs against the
    // JA-default browser. No step suffix: every Sir Tommy foundation builds +1.
    await expect(page.getByLabel(/ファンデーション 0 /).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByLabel(/ファンデーション 3 /).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Control buttons while playing
    await expect(page.getByRole('button', { name: 'ヒント' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button targeted via data-tutorial
    await expect(page.locator('[data-tutorial="ca-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/sirtommy');
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();

    // After giveup, the action log button should appear
    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
