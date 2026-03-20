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
    const suitSpade = page.getByRole('button', { name: '♠' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let sawPlay = false;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(playButton.or(drawButton).or(nextRoundButton).or(suitSpade).first()).toBeVisible({
        timeout: 10_000,
      });

      const playVisible = await playButton.isVisible().catch(() => false);
      const drawVisible = await drawButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);
      const suitVisible = await suitSpade.isVisible().catch(() => false);

      // Game end: no action buttons visible
      if (!playVisible && !drawVisible && !nextRoundVisible && !suitVisible) break;

      // Suit choice phase: pick spade
      if (suitVisible) {
        await suitSpade.click();
        await waitForLoaded(page);
        continue;
      }

      // Play phase: select a card and play, or draw
      if (playVisible || drawVisible) {
        sawPlay = true;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible().catch(() => false)) && (await playButton.isEnabled())) {
          await playButton.click();
          await waitForLoaded(page);
          continue;
        }
        // If play not possible, draw
        if ((await drawButton.isVisible().catch(() => false)) && (await drawButton.isEnabled())) {
          await drawButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    // Verify we saw play phase
    expect(sawPlay).toBe(true);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
