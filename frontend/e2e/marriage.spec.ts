import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Marriage E2E', () => {
  test('navigates, resets, and completes a draw/discard interaction', async ({ page }) => {
    await navigateTo(page, '/marriage');

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
    const handCards = page.locator('button[aria-pressed]:has(img)');

    await expect(drawStockButton.or(drawDiscardButton).or(discardButton).first()).toBeVisible({ timeout: 10_000 });

    if (await drawStockButton.isVisible()) {
      await drawStockButton.click();
    } else if (await drawDiscardButton.isVisible()) {
      await drawDiscardButton.click();
    } else {
      const cardCount = await handCards.count();
      if (cardCount > 0) await handCards.first().click();
      await expect(discardButton).toBeEnabled();
      await discardButton.click();
    }
    await waitForLoaded(page);
  });
});
