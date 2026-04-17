import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Scorpion E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/Stock|ストック/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset button restarts the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    // Wait for game state to fully load (reset API call returns, tableau renders)
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await waitForLoaded(page);
  });

  test('deal button is visible', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const dealButton = page.getByRole('button', { name: '配る' }).first();
    await expect(dealButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' }).first();
    await expect(giveUpButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await giveUpButton.click();
    await waitForLoaded(page);
    await expect(page.getByText('ゲームオーバー').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
