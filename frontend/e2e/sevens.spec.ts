import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Sevens E2E', () => {
  test('plays a full game: reset → play or pass each turn → end → reset', async ({ page }) => {
    await navigateTo(page, '/sevens');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Game loop
    for (let turn = 0; turn < 300; turn++) {
      const passButton = page.getByRole('button', { name: 'パス' });

      // Check if game ended: all players finished
      const gameEnd = page.locator('text=ゲーム終了');
      if (await gameEnd.isVisible({ timeout: 1_000 }).catch(() => false)) {
        break;
      }

      // Try to find a playable card (green-bordered) and click it
      const playableCards = page.locator('.border-green-400, .border-green-500, [class*="border-green"]');
      const playableCount = await playableCards.count();

      if (playableCount > 0) {
        await playableCards.first().click();
        await waitForLoaded(page);
      } else if (await passButton.isVisible({ timeout: 2_000 }).catch(() => false)) {
        if (await passButton.isEnabled()) {
          await passButton.click();
          await waitForLoaded(page);
        }
      }

      await page.waitForTimeout(200);
    }

    // Reset for next game
    await resetButton.click();
    await waitForLoaded(page);
  });
});
