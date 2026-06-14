import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Gaps E2E', () => {
  test('navigates, renders the grid, and supports hint + reset', async ({ page }) => {
    await navigateTo(page, '/gaps');

    // Header pieces are visible.
    await expect(page.getByText(/手数/)).toBeVisible();
    await expect(page.getByText(/再配り残り/)).toBeVisible();

    // Hint
    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    // Reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('redeal button decrements the remaining counter', async ({ page }) => {
    await navigateTo(page, '/gaps');
    const redealButton = page.getByRole('button', { name: /再配り/ });
    await expect(redealButton).toBeVisible();
    await redealButton.click();
    await waitForLoaded(page);
    // After one redeal, the counter should reflect 1/3.
    await expect(page.getByText(/1\/3/)).toBeVisible();
  });

  test('give up ends the game and hides action buttons', async ({ page }) => {
    await navigateTo(page, '/gaps');
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
