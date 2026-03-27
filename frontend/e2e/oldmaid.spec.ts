import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Old Maid E2E', () => {
  // Old Maid has 800ms animation delays per CPU turn, so a full game can take minutes
  test('starts a game and draws a card', async ({ page }) => {
    await navigateTo(page, '/oldmaid');

    // Game auto-starts with default settings; wait for it to load
    await waitForLoaded(page);

    // Wait for the game to initialize and reach human's turn
    const randomDrawButton = page.getByRole('button', { name: 'ランダムに引く' });
    await expect(randomDrawButton).toBeVisible({ timeout: 30_000 });

    // Draw a card
    await randomDrawButton.click();
    await waitForLoaded(page);

    // After drawing, the game processes CPU turns (with animation delays)
    // Wait until either the draw button reappears (human's turn again) or the game ends
    const resetButton = page.getByRole('button', { name: 'リセット' });
    const nextAction = randomDrawButton.or(resetButton);
    await expect(nextAction.first()).toBeVisible({ timeout: 30_000 });
  });
});
