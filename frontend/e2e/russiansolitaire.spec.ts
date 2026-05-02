import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Russian Solitaire E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/russiansolitaire');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset button is present', async ({ page }) => {
    await navigateTo(page, '/russiansolitaire');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="rs-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/russiansolitaire');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' }).first();
    await expect(giveUpButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await giveUpButton.click();
    await expect(page.getByText('ゲームオーバー').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
