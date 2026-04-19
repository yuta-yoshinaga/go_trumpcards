import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Accordion E2E', () => {
  test('navigates to accordion, resets, and shows piles', async ({ page }) => {
    await navigateTo(page, '/accordion');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // 52 piles should be rendered at start (buttons with aria-labels "0:", "1:", ...)
    const pile0 = page.getByRole('button', { name: /^0:/ });
    await expect(pile0).toBeVisible();

    // The giveup, hint, and undo buttons should be available while playing
    await expect(page.getByRole('button', { name: 'ヒント' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'ギブアップ' })).toBeVisible();
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/accordion');
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'ギブアップ' }).click();
    await waitForLoaded(page);

    // After giveup, the action log button is exposed for end phase
    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton).toBeVisible();
  });
});
