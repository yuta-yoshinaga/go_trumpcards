import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Scorpion E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    await expect(page.getByText(/手数/).first()).toBeVisible();
    await expect(page.getByText(/Stock|ストック/).first()).toBeVisible();
  });

  test('reset button restarts the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);
  });

  test('deal button is visible', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    const dealButton = page.getByRole('button', { name: '配る' }).first();
    await expect(dealButton).toBeVisible();
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/scorpion');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' }).first();
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();
    await waitForLoaded(page);
    await expect(page.getByText('ゲームオーバー').first()).toBeVisible();
  });
});
