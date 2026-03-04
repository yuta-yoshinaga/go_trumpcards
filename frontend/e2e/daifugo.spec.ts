import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Daifugo E2E', () => {
  test('plays a full game: reset → pass repeatedly until end → reset', async ({ page }) => {
    await navigateTo(page, '/daifugo');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Game loop: pass on every turn until game ends
    let gameEnded = false;
    for (let turn = 0; turn < 300; turn++) {
      const passButton = page.getByRole('button', { name: 'パス' });

      // Check if game ended (player has a rank displayed)
      const finishedText = page.locator('text=大富豪').or(page.locator('text=大貧民'));
      if (
        await finishedText
          .first()
          .isVisible({ timeout: 1_000 })
          .catch(() => false)
      ) {
        // Check if pass button is gone (game fully ended)
        if (!(await passButton.isVisible().catch(() => false))) {
          gameEnded = true;
          break;
        }
      }

      // Pass if the button is available
      if (await passButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
        if (await passButton.isEnabled()) {
          await passButton.click();
          await waitForLoaded(page);
        }
      }

      await waitForLoaded(page);
    }

    expect(gameEnded).toBe(true);

    // Reset for next game
    await resetButton.click();
    await waitForLoaded(page);
  });
});
