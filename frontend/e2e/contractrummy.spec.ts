import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Contract Rummy E2E', () => {
  test('navigates and shows the round contract banner', async ({ page }) => {
    await navigateTo(page, '/contractrummy');

    // Round / contract banner is visible
    await expect(page.getByText(/ラウンド 1 \/ 7|Round 1 \/ 7/)).toBeVisible();
    await expect(page.getByText(/今回のコントラクト|Contract/)).toBeVisible();
  });

  test('lets the human draw from the stock', async ({ page }) => {
    await navigateTo(page, '/contractrummy');

    const drawButton = page.getByRole('button', { name: /山札から引く|Draw from stock/ });
    await expect(drawButton).toBeVisible();
    await drawButton.click();
    await waitForLoaded(page);

    // After drawing, a discard button should be available (play phase).
    await expect(page.getByRole('button', { name: /カードを捨てる|Discard card/ })).toBeVisible();
  });

  test('reset returns the page to draw-phase controls', async ({ page }) => {
    await navigateTo(page, '/contractrummy');

    const resetButton = page.getByRole('button', { name: /リセット|Reset/ });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    // Confirmation dialog appears for mid-game resets.
    const confirm = page.getByRole('button', { name: /確認|Confirm/ });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: /山札から引く|Draw from stock/ })).toBeVisible();
  });
});
