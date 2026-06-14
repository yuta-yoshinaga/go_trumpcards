import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe("Baker's Dozen E2E", () => {
  test('navigates, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/bakersdozen');

    // Verify move count label is displayed
    await expect(page.getByText(/手数/)).toBeVisible();

    // Click hint button
    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    // Click reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game restarted (move count visible again)
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/bakersdozen');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
