import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Pyramid E2E', () => {
  test('navigates, resets, and plays basic moves', async ({ page }) => {
    await navigateTo(page, '/pyramid');

    // Verify stock and waste labels are visible
    await expect(page.getByText('山札')).toBeVisible();
    await expect(page.getByText('ウェイスト')).toBeVisible();

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
    await navigateTo(page, '/pyramid');

    // Click give up
    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
