import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Braid E2E', () => {
  test('fixes the direction, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/braid');

    await expect(page.getByText(/手数/)).toBeVisible();
    // Nothing reaches a foundation until the direction is fixed.
    await expect(page.getByText(/向き: 未選択/)).toBeVisible();

    await page.getByTestId('direction-up').click();
    await waitForLoaded(page);
    await expect(page.getByText(/向き: 昇順/)).toBeVisible();
    await expect(page.getByTestId('direction-up')).toHaveCount(0);

    const drawButton = page.getByRole('button', { name: 'めくる', exact: true });
    await expect(drawButton).toBeVisible();
    await drawButton.click();
    await waitForLoaded(page);

    const hintButton = page.getByRole('button', { name: 'ヒント', exact: true });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // A reset deals a fresh game, so the direction is open again.
    await expect(page.getByText(/向き: 未選択/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/braid');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント', exact: true })).not.toBeVisible();
  });
});
