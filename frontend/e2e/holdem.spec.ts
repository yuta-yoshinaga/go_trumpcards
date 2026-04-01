import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_QUICK, waitForLoaded } from './helpers';

test.describe("Texas Hold'em E2E", () => {
  test('plays a full round: reset → check/call through rounds → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/holdem');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through betting rounds: PRE_FLOP → FLOP → TURN → RIVER → SHOWDOWN
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });

      // Check if we've reached the end
      if (await isVisibleWithin(resetButton, TIMEOUT_QUICK)) {
        const checkVisible = await checkButton.isVisible();
        const callVisible = await callButton.isVisible();
        if (!checkVisible && !callVisible) {
          roundEnded = true;
          break;
        }
      }

      // Try check first, then call
      if (await isVisibleWithin(checkButton, TIMEOUT_ACTION)) {
        if (await checkButton.isEnabled()) {
          await checkButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      if (await isVisibleWithin(callButton, TIMEOUT_ACTION)) {
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
