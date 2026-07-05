import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Anaconda E2E', () => {
  // Anaconda (Pass the Trash) is a poker pot game: the human passes cards left,
  // keeps their best 5, then bets between reveals during the Roll phase.
  test('plays through the pass phase', async ({ page }) => {
    await navigateTo(page, '/anaconda');

    // The pass phase shows the Pass button (disabled until cards are selected).
    const passButton = page.getByRole('button', { name: /Pass|パスする/ });
    await expect(passButton.first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    // Select the required cards by clicking the first few hand cards, then pass.
    const handCards = page.locator('button[aria-pressed]');
    await expect(handCards.first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    await handCards.nth(0).click();
    await handCards.nth(1).click();
    await handCards.nth(2).click();

    if (await isVisibleWithin(passButton.first(), TIMEOUT_ACTION)) {
      await passButton.first().click();
      await waitForLoaded(page);
    }

    // The game should still be playable (some action control remains visible).
    await expect(
      page.getByRole('button', { name: /Pass|パスする|Keep|キープ|Call|コール|Reset|リセット/ }).first(),
    ).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
