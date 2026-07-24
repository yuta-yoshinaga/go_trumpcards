import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Tonk E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/tonk');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const drawDiscardButton = page.getByRole('button', { name: '捨て札から引く', exact: true });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const knockButton = page.getByRole('button', { name: 'ノック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        drawStockButton
          .or(drawDiscardButton)
          .or(discardButton)
          .or(knockButton)
          .or(nextRoundButton)
          .or(anyResetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      const drawStockVisible = await drawStockButton.isVisible();
      const drawDiscardVisible = await drawDiscardButton.isVisible();
      const discardVisible = await discardButton.isVisible();
      const knockVisible = await knockButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      if (!drawStockVisible && !drawDiscardVisible && !discardVisible && !knockVisible && !nextRoundVisible) break;

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

      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    expect(interactions).toBeGreaterThan(0);

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
