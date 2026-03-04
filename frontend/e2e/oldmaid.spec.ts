import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Old Maid E2E', () => {
  test('plays a full game: start → draw cards until end → reset', async ({ page }) => {
    await navigateTo(page, '/oldmaid');

    // Setup screen: click ゲーム開始
    const startButton = page.getByRole('button', { name: 'ゲーム開始' });
    await expect(startButton).toBeVisible();
    await startButton.click();
    await waitForLoaded(page);

    // Game loop: keep drawing until finished
    let gameEnded = false;
    for (let turn = 0; turn < 200; turn++) {
      const randomDrawButton = page.getByRole('button', { name: 'ランダムに引く' });
      const resetButton = page.getByRole('button', { name: 'リセット' });

      // Check if game has ended (リセット is visible, ランダムに引く is not)
      if (await resetButton.isVisible({ timeout: 1_000 }).catch(() => false)) {
        const drawVisible = await randomDrawButton.isVisible().catch(() => false);
        if (!drawVisible) {
          gameEnded = true;
          // Game has ended
          await resetButton.click();
          await waitForLoaded(page);
          // Back to setup screen
          await expect(startButton).toBeVisible();
          break;
        }
      }

      // If it's human's turn, draw randomly
      if (await randomDrawButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
        await randomDrawButton.click();
        await waitForLoaded(page);
      }

      // Wait for CPU turns to process
      await waitForLoaded(page);
    }

    expect(gameEnded).toBe(true);
  });
});
