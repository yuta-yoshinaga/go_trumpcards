import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

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
      const gameEnd = page.locator('text=ゲーム終了');
      const playableCards = page.locator('[data-testid="playable-card"]');

      // Wait for whichever of the three appears first. Probing the end state
      // with its own timeout cost a full second every turn even mid-game.
      await expect(gameEnd.or(playableCards).or(passButton).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

      if (await gameEnd.isVisible()) {
        gameEnded = true;
        break;
      }

      if ((await playableCards.count()) > 0) {
        await playableCards.first().click();
      } else if (await passButton.isEnabled()) {
        await passButton.click();
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
