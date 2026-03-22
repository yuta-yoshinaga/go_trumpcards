import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Indian Poker E2E', () => {
  test('plays a full round: reset → bet/check through betting → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/indianpoker');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through betting round
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック' });
      const callButton = page.getByRole('button', { name: 'コール' });

      // Check if we've reached the end
      if (await resetButton.isVisible({ timeout: 1_000 }).catch(() => false)) {
        const checkVisible = await checkButton.isVisible().catch(() => false);
        const callVisible = await callButton.isVisible().catch(() => false);
        if (!checkVisible && !callVisible) {
          roundEnded = true;
          break;
        }
      }

      // Try check first, then call
      if (await checkButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
        if (await checkButton.isEnabled()) {
          await checkButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      if (await callButton.isVisible({ timeout: 2_000 }).catch(() => false)) {
        if (await callButton.isEnabled()) {
          await callButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      await waitForLoaded(page);
    }

    expect(roundEnded).toBe(true);

    // Start next round
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
  });
});
