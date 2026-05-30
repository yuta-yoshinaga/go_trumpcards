import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Aces Up E2E', () => {
  test('navigates, deals, and plays basic moves', async ({ page }) => {
    await navigateTo(page, '/acesup');

    // Verify stock label and move count are visible
    await expect(page.getByText('山札')).toBeVisible();
    await expect(page.getByText(/手数/)).toBeVisible();

    // Deal a card to each column
    const dealButton = page.getByRole('button', { name: '配る' }).last();
    await expect(dealButton).toBeVisible();
    await dealButton.click();
    await waitForLoaded(page);

    // Deal a few more times
    for (let i = 0; i < 3; i++) {
      await dealButton.click();
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
    await navigateTo(page, '/acesup');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
