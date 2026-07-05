import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Indian Rummy E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/indianrummy');

    // Start (mid-game reset -> confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+\/\d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const drawDiscardButton = page.getByRole('button', { name: '捨て札から引く' });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        drawStockButton.or(drawDiscardButton).or(discardButton).or(nextRoundButton).or(anyResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const drawStockVisible = await drawStockButton.isVisible();
      const drawDiscardVisible = await drawDiscardButton.isVisible();
      const discardVisible = await discardButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Game end: no action buttons visible
      if (!drawStockVisible && !drawDiscardVisible && !discardVisible && !nextRoundVisible) break;

      // Draw phase
      if (drawStockVisible || drawDiscardVisible) {
        interactions++;
        if (drawStockVisible) {
          await drawStockButton.click();
        } else {
          await drawDiscardButton.click();
        }
        await waitForLoaded(page);
        continue;
      }

      // Discard phase: select a card and discard (never auto-declare — validity unknown)
      if (discardVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) await handCards.first().click();
        if (await discardButton.isEnabled()) {
          await discardButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    expect(interactions).toBeGreaterThan(0);

    // Reset and verify restart
    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+\/\d+$/).first()).toBeVisible();
  });
});
