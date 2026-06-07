import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Easthaven E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/easthaven');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/Stock|ストック/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset button is present', async ({ page }) => {
    await navigateTo(page, '/easthaven');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="eh-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('deal button is visible', async ({ page }) => {
    await navigateTo(page, '/easthaven');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '配る' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/easthaven');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' }).first();
    await expect(giveUpButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await giveUpButton.click();
    await expect(page.getByText('ゲームオーバー').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
