import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Poker E2E', () => {
  test('plays a full round: reset → bet → stand (no exchange) → bet → result → reset', async ({ page }) => {
    await navigateTo(page, '/poker');

    // Click リセット to start a new game
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // DEAL phase: use チェック or コール to proceed
    const checkButton = page.getByRole('button', { name: 'チェック' });
    const callButton = page.getByRole('button', { name: 'コール' });
    if (await checkButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await checkButton.click();
    } else {
      await callButton.click();
    }
    await waitForLoaded(page);

    // EXCHANGE phase: click スタンド (no exchange)
    const standButton = page.getByRole('button', { name: 'スタンド' });
    if (await standButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await standButton.click();
      await waitForLoaded(page);
    }

    // SECOND_BET phase: チェック or コール
    if (await checkButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await checkButton.click();
    } else if (await callButton.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await callButton.click();
    }
    await waitForLoaded(page);

    // END phase: リセット should be visible again
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Start another round
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
  });
});
