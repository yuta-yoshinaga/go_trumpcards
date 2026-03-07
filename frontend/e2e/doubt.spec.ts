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

    const MAX_TURNS = 300;
    const gameEnd = page.locator('text=ゲーム終了');
    const skipButton = page.getByRole('button', { name: 'スルー' });
    const confirmButton = page.getByRole('button', { name: '確認' });
    const playButton = page.getByRole('button', { name: '出す' });
    const handCards = page.locator('[data-testid="hand-card"]');

    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Wait for any actionable element or game end to appear
      await expect(gameEnd.or(skipButton).or(confirmButton).or(playButton)).toBeVisible({ timeout: 10_000 });

      // Check if game ended (instant, no timeout)
      if (await gameEnd.isVisible()) break;

      // Doubt phase: skip doubt or confirm CPU doubt
      if (await skipButton.isVisible()) {
        await skipButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (await confirmButton.isVisible()) {
        await confirmButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Play phase: select first card and play
      if (await playButton.isVisible()) {
        if ((await handCards.count()) > 0) {
          await handCards.first().click();
        }
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
        }
      }
    }

    // Assert game ended (Playwright auto-retry)
    await expect(gameEnd).toBeVisible({ timeout: 5_000 });

    // Reset
    await resetButton.click();
    await waitForLoaded(page);
  });
});
