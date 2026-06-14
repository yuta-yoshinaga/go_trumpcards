import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Spiderette E2E', () => {
  test('navigates, resets, and shows basic controls', async ({ page }) => {
    await navigateTo(page, '/spiderette');

    await expect(page.getByText('山札')).toBeVisible();
    await expect(page.getByText(/手数/)).toBeVisible();

    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await expect(page.getByRole('button', { name: '自動完成' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).toBeVisible();
    await expect(page.getByRole('button', { name: '配る' }).last()).toBeVisible();

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/spiderette');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // After give up, the playing buttons should be gone
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
