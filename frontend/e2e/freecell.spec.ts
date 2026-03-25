import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('FreeCell E2E', () => {
  test('navigates, resets, and plays basic moves', async ({ page }) => {
    await navigateTo(page, '/freecell');

    // Verify foundation suit symbols (scoped to game area to exclude nav icons)
    const gameArea = page.locator('[aria-busy]');
    await expect(gameArea.getByText('♠', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♣', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♥', { exact: true })).toBeVisible();
    await expect(gameArea.getByText('♦', { exact: true })).toBeVisible();

    // Verify move count is displayed
    await expect(page.getByText(/手数/)).toBeVisible();

    // Verify playing buttons are visible
    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await expect(page.getByRole('button', { name: 'オートコンプリート' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).toBeVisible();

    // Click reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game restarted
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/freecell');

    // Click give up
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
    await expect(page.getByRole('button', { name: 'オートコンプリート' })).not.toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).not.toBeVisible();

    // Reset should still be visible
    await expect(page.getByRole('button', { name: 'リセット' })).toBeVisible();

    // Action log button should appear
    await expect(page.getByRole('button', { name: '棋譜を見る' })).toBeVisible();

    // Reset to start a new game
    await page.getByRole('button', { name: 'リセット' }).click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ヒント' })).toBeVisible();
  });
});
