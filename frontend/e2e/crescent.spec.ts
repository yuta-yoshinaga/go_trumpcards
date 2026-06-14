import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Crescent Solitaire E2E', () => {
  test('navigates, displays redeal count, and resets', async ({ page }) => {
    await navigateTo(page, '/crescent');

    await expect(page.getByText(/残り再配り回数: 3/)).toBeVisible();
    await expect(page.getByText(/手数/)).toBeVisible();

    const redealButton = page.getByRole('button', { name: /再配り \(3\)/ });
    await expect(redealButton).toBeVisible();
    await redealButton.click();
    await waitForLoaded(page);

    // Counter should drop after a redeal.
    await expect(page.getByText(/残り再配り回数: 2/)).toBeVisible();

    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/残り再配り回数: 3/)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/crescent');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
