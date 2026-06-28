import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Prší E2E', () => {
  test('navigates, resets, and plays through the hand', async ({ page }) => {
    await navigateTo(page, '/prsi');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify draw-pile info is visible
    await expect(page.getByText(/^山札: \d+枚$/).first()).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const drawButton = page.getByRole('button', { name: '引く' });
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify the hand progresses.
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Break cleanly once the game reaches end state (only 次のゲーム remains).
      if (await endResetButton.isVisible()) break;

      const playVisible = await isVisibleWithin(playButton, TIMEOUT_GAME_LOOP);
      const drawVisible = await drawButton.isVisible();

      // Not the human's turn (CPUs auto-play); wait briefly and re-check.
      if (!playVisible && !drawVisible) {
        if (await endResetButton.isVisible()) break;
        await page.waitForTimeout(300);
        continue;
      }

      interactions++;
      const cardCount = await handCards.count();
      if (cardCount > 0) {
        await handCards.first().click();
      }
      if ((await playButton.isVisible()) && (await playButton.isEnabled())) {
        await playButton.click();
        await waitForLoaded(page);
        continue;
      }
      // If play is not possible, draw.
      if ((await drawButton.isVisible()) && (await drawButton.isEnabled())) {
        await drawButton.click();
        await waitForLoaded(page);
      }
    }

    // Verify we had at least one interaction (play or draw).
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts. Button could be mid-game (リセット) or end (次のゲーム).
    if (await midResetButton.isVisible()) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^山札: \d+枚$/).first()).toBeVisible();
  });
});
