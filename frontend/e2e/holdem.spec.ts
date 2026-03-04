import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe("Texas Hold'em E2E", () => {
  test('plays a full round: reset → check/call through rounds → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/holdem');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Play through betting rounds: PRE_FLOP → FLOP → TURN → RIVER → SHOWDOWN
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック' });
      const callButton = page.getByRole('button', { name: 'コール' });

      // Check if we've reached the end
      if (await resetButton.isVisible({ timeout: 1_000 }).catch(() => false)) {
        const checkVisible = await checkButton.isVisible().catch(() => false);
        const callVisible = await callButton.isVisible().catch(() => false);
        if (!checkVisible && !callVisible) {
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

      await page.waitForTimeout(300);
    }

    // Start next round
    await resetButton.click();
    await waitForLoaded(page);
  });
});
