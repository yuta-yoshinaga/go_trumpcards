import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Doubt E2E', () => {
  test('plays a full game: reset → select card + play / skip doubt → end → reset', async ({ page }) => {
    await navigateTo(page, '/doubt');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Game loop
    for (let turn = 0; turn < 200; turn++) {
      // Check if game ended
      const gameEnd = page.locator('text=ゲーム終了');
      if (await gameEnd.isVisible({ timeout: 1_000 }).catch(() => false)) {
        break;
      }

      // Doubt phase: skip doubt or confirm
      const skipButton = page.getByRole('button', { name: 'スルー' });
      const confirmButton = page.getByRole('button', { name: '確認' });

      if (await skipButton.isVisible({ timeout: 1_000 }).catch(() => false)) {
        await skipButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (await confirmButton.isVisible({ timeout: 1_000 }).catch(() => false)) {
        await confirmButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Play phase: select first card and play
      const playButton = page.getByRole('button', { name: '出す' });
      if (await playButton.isVisible({ timeout: 2_000 }).catch(() => false)) {
        // Select the first card in hand
        const handCards = page
          .locator('[aria-busy="false"] button[class*="border"]')
          .or(page.locator('img[alt]').first());
        if (
          await handCards
            .first()
            .isVisible()
            .catch(() => false)
        ) {
          await handCards.first().click();
        }

        // Click 出す if enabled
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
        }
      }

      await page.waitForTimeout(500);
    }

    // Reset
    await resetButton.click();
    await waitForLoaded(page);
  });
});
