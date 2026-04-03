import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Crazy Eights E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/crazyeights');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const drawButton = page.getByRole('button', { name: '引く' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const suitSpade = page.getByRole('button', { name: '♠', exact: true });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(playButton.or(drawButton).or(nextRoundButton).or(suitSpade).first()).toBeVisible({
        timeout: 10_000,
      });

      const playVisible = await playButton.isVisible();
      const drawVisible = await drawButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();
      const suitVisible = await suitSpade.isVisible();

      // Game end: no action buttons visible
      if (!playVisible && !drawVisible && !nextRoundVisible && !suitVisible) break;

      // Suit choice phase: pick spade
      if (suitVisible) {
        await suitSpade.click();
        await waitForLoaded(page);
        interactions++;
        continue;
      }

      // Play phase: select a card and play, or draw
      if (playVisible || drawVisible) {
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
        // If play not possible, draw
        if ((await drawButton.isVisible()) && (await drawButton.isEnabled())) {
          await drawButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        await nextRoundButton.click();
        await waitForLoaded(page);
        interactions++;
      }
    }

    // Verify we had at least one interaction (play, draw, suit choice, or round end)
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
