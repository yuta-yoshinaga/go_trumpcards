import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Seven Card Stud E2E', () => {
  test('plays a full round: reset → check/call through streets → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/sevencardstud');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through streets: THIRD_STREET → FOURTH → FIFTH → SIXTH → SEVENTH → SHOWDOWN
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });

      // Check if we've reached the end
      if (await endResetButton.isVisible()) {
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

    // Start next round (end state: no confirm dialog)
    await endResetButton.click();
    await waitForLoaded(page);
  });
});
