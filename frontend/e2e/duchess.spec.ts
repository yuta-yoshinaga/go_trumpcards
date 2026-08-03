import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Duchess E2E', () => {
  test('picks the base rank, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/duchess');

    // Nothing else is legal until the base rank is set, so the board opens
    // with it unset and drawing disabled.
    await expect(page.getByText(/開始ランク: 未決定/)).toBeVisible();
    await expect(page.getByRole('button', { name: 'めくる', exact: true })).toBeDisabled();

    await page.getByRole('button', { name: 'リザーブ扇 0 の札を開始ランクにする' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/開始ランク: 未決定/)).not.toBeVisible();

    const drawButton = page.getByRole('button', { name: 'めくる', exact: true });
    await expect(drawButton).toBeEnabled();
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

    await expect(page.getByText(/開始ランク: 未決定/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/duchess');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント', exact: true })).not.toBeVisible();
  });
});
