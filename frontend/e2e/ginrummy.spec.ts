import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Gin Rummy E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/ginrummy');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア').first()).toBeVisible();

    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const drawDiscardButton = page.getByRole('button', { name: '捨て札から引く' });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const layoffButton = page.getByRole('button', { name: 'レイオフ' });
    const skipButton = page.getByRole('button', { name: 'スキップ' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let sawDraw = false;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        drawStockButton.or(discardButton).or(layoffButton).or(skipButton).or(nextRoundButton).or(resetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const drawStockVisible = await drawStockButton.isVisible().catch(() => false);
      const drawDiscardVisible = await drawDiscardButton.isVisible().catch(() => false);
      const discardVisible = await discardButton.isVisible().catch(() => false);
      const layoffVisible = await layoffButton.isVisible().catch(() => false);
      const skipVisible = await skipButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);

      // Game end: no action buttons visible
      if (
        !drawStockVisible &&
        !drawDiscardVisible &&
        !discardVisible &&
        !layoffVisible &&
        !skipVisible &&
        !nextRoundVisible
      )
        break;

      // Draw phase: draw from stock or discard pile
      if (drawStockVisible) {
        sawDraw = true;
        await drawStockButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Discard phase: select a card and discard (or knock)
      if (discardVisible) {
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if (await discardButton.isEnabled()) {
          await discardButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Layoff phase: skip layoff
      if (layoffVisible || skipVisible) {
        if (skipVisible) {
          await skipButton.click();
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

    // Verify we saw draw phase
    expect(sawDraw).toBe(true);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
