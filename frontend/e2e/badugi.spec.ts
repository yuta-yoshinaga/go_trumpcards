import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Badugi E2E', () => {
  test('plays a full hand: reset → check/call + stand → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/badugi');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through up to 4 betting rounds + 3 draw rounds with generous retry.
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });
      const standButton = page.getByRole('button', { name: 'スタンド', exact: true });

      if (await endResetButton.isVisible()) {
        const checkVisible = await checkButton.isVisible();
        const callVisible = await callButton.isVisible();
        const standVisible = await standButton.isVisible();
        if (!checkVisible && !callVisible && !standVisible) {
          roundEnded = true;
          break;
        }
      }

      // Draw phase has Stand button visible — click it (no exchanges).
      if (await isVisibleWithin(standButton, TIMEOUT_ACTION)) {
        if (await standButton.isEnabled()) {
          await standButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      // Betting phase: prefer check, fall back to call.
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

    await endResetButton.click();
    await waitForLoaded(page);
  });
});
