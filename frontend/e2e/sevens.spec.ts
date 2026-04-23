import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_QUICK, waitForLoaded } from './helpers';

test.describe('Sevens E2E', () => {
  test('plays a full game: reset → play or pass each turn → end → reset', async ({ page }) => {
    await navigateTo(page, '/sevens');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Game loop
    let gameEnded = false;
    for (let turn = 0; turn < 300; turn++) {
      const passButton = page.getByRole('button', { name: 'パス' });

      // Check if game ended: all players finished
      const gameEnd = page.locator('text=ゲーム終了');
      if (await isVisibleWithin(gameEnd, TIMEOUT_QUICK)) {
        gameEnded = true;
        break;
      }

      // Try to find a playable card via data-testid and click it
      const playableCards = page.locator('[data-testid="playable-card"]');
      const playableCount = await playableCards.count();

      if (playableCount > 0) {
        await playableCards.first().click();
        await waitForLoaded(page);
      } else if (await isVisibleWithin(passButton, TIMEOUT_ACTION)) {
        if (await passButton.isEnabled()) {
          await passButton.click();
          await waitForLoaded(page);
        }
      }

      await waitForLoaded(page);
    }

    expect(gameEnded).toBe(true);

    // Reset for next game (end state: no confirm dialog)
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    await endResetButton.click();
    await waitForLoaded(page);
  });
});
