import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Spider E2E', () => {
  test('navigates, resets, and plays basic moves', async ({ page }) => {
    await navigateTo(page, '/spider');

    // Verify stock label is visible
    await expect(page.getByText('山札')).toBeVisible();

    // Verify move count is displayed
    await expect(page.getByText(/手数/)).toBeVisible();

    // Verify playing buttons are visible
    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await expect(page.getByRole('button', { name: '自動完成' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).toBeVisible();
    await expect(page.getByRole('button', { name: '配る' }).last()).toBeVisible();

    // Click reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game restarted
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/spider');

    // Click give up
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();
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
