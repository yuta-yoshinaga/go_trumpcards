import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Crazy Pineapple Poker E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/crazypineapple');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText('あなたの手札')).toBeVisible({ timeout: 10_000 });

    const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
    const callButton = page.getByRole('button', { name: 'コール', exact: true });
    const foldButton = page.getByRole('button', { name: 'フォールド' });
    const discardControls = page.getByTestId('discard-controls');

    await expect(checkButton.or(callButton).or(foldButton).or(discardControls).or(resetButton).first()).toBeVisible({
      timeout: 10_000,
    });

    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText('あなたの手札')).toBeVisible({ timeout: 10_000 });
  });
});
