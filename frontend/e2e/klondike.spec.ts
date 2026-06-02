import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Klondike E2E', () => {
  test('navigates, resets, and plays basic moves', async ({ page }) => {
    await navigateTo(page, '/klondike');

    // Verify stock and waste labels are visible
    await expect(page.getByText('山札')).toBeVisible();
    await expect(page.getByText('ウェイスト')).toBeVisible();

    // Verify foundation suit symbols (scoped to game area to exclude nav icons)
    const gameArea = page.locator('[aria-busy]');
    await expect(gameArea.getByText('♠', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♣', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♥', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♦', { exact: true })).toBeVisible();

    // Verify move count is displayed
    await expect(page.getByText(/手数/)).toBeVisible();

    // Click draw button to draw from stock
    const drawButton = page.getByRole('button', { name: '引く' }).last();
    await expect(drawButton).toBeVisible();
    await drawButton.click();
    await waitForLoaded(page);

    // Draw a few more times
    for (let i = 0; i < 3; i++) {
      await drawButton.click();
      await waitForLoaded(page);
    }

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

    // Verify game restarted
    await expect(page.getByText('山札')).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/klondike');

    // Click give up — gated behind a confirm dialog since #2099.
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog.
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
    await expect(page.getByRole('button', { name: '自動完成' })).not.toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).not.toBeVisible();

    // 次のゲーム button (end state) should be visible
    await expect(page.getByRole('button', { name: '次のゲーム' })).toBeVisible();

    // Action log button should appear
    await expect(page.getByRole('button', { name: '棋譜を見る' })).toBeVisible();

    // Start a new game (end state: no confirm dialog)
    await page.getByRole('button', { name: '次のゲーム' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ヒント' })).toBeVisible();
  });
});
