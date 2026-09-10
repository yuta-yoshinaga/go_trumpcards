import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Biriba E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/biriba');

    const resetButton = page.getByRole('button', { name: /リセット|Reset/ });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    const confirmButton = page.getByRole('button', { name: /確認|Confirm/ });
    if (await confirmButton.isVisible()) await confirmButton.click();
    await waitForLoaded(page);

    await expect(page.getByText(/ラウンド \d+|Round \d+/).first()).toBeVisible();
    const drawStock = page.getByRole('button', { name: /山札から引く|Draw from stock/ });
    const skipMeld = page.getByRole('button', { name: /スキップ|Skip/ });
    const discard = page.getByRole('button', { name: /捨てる|Discard/ });
    const nextRound = page.getByRole('button', { name: /次のラウンド|Next round/ });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    let interactions = 0;
    for (let turn = 0; turn < 60; turn++) {
      if (await drawStock.isVisible()) {
        interactions++;
        await drawStock.click();
      } else if (await skipMeld.isVisible()) {
        interactions++;
        await skipMeld.click();
      } else if (await discard.isVisible()) {
        interactions++;
        await handCards.first().click();
        await discard.click();
      } else if (await nextRound.isVisible()) {
        interactions++;
        await nextRound.click();
      } else {
        break;
      }
      await waitForLoaded(page);
    }

    expect(interactions).toBeGreaterThan(0);
  });
});
