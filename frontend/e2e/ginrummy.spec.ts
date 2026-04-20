import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Gin Rummy E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/ginrummy');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const drawDiscardButton = page.getByRole('button', { name: '捨て札から引く' });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const layoffButton = page.getByRole('button', { name: 'レイオフ' });
    const skipButton = page.getByRole('button', { name: 'スキップ' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        drawStockButton
          .or(drawDiscardButton)
          .or(discardButton)
          .or(layoffButton)
          .or(skipButton)
          .or(nextRoundButton)
          .or(anyResetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      const drawStockVisible = await drawStockButton.isVisible();
      const drawDiscardVisible = await drawDiscardButton.isVisible();
      const discardVisible = await discardButton.isVisible();
      const layoffVisible = await layoffButton.isVisible();
      const skipVisible = await skipButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

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

      // Discard phase: select a card and discard (or knock)
      if (discardVisible) {
        interactions++;
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

      // Layoff phase: lay off a card or skip
      if (layoffVisible || skipVisible) {
        interactions++;
        if (layoffVisible) {
          const cardCount = await handCards.count();
          if (cardCount > 0) await handCards.first().click();
          if ((await layoffButton.isVisible()) && (await layoffButton.isEnabled())) {
            await layoffButton.click();
            await waitForLoaded(page);
            continue;
          }
        }
        // Fall through to skip if layoff is not possible
        if (skipVisible || (await skipButton.isVisible())) {
          await skipButton.click();
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

    // Verify we had at least one interaction
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts. Button could be mid-game (リセット) or end (次のゲーム).
    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
