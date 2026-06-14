import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Scorpion E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/Stock|ストック/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset button is present', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    // Wait for game state to fully load before asserting footer buttons
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Target via data-tutorial attribute to avoid collision with NavBar / other "リセット" labels
    await expect(page.locator('[data-tutorial="sc-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('deal button is visible', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '配る' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' }).first();
    await expect(giveUpButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await expect(page.getByText('ゲームオーバー').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
