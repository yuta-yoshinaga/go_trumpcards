import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Anaconda E2E', () => {
  // Anaconda (Pass the Trash) is a poker pot game: the human passes cards left,
  // keeps their best 5, then bets between reveals during the Roll phase.
  test('renders the pass phase and toggles card selection', async ({ page }) => {
    await navigateTo(page, '/anaconda');
    await waitForLoaded(page);

    // The pass phase shows the Pass button and the human's selectable hand.
    await expect(page.getByRole('button', { name: /Pass|パスする/ }).first()).toBeVisible({
      timeout: TIMEOUT_ACTION,
    });
    const handCards = page.locator('button[aria-pressed]');
    await expect(handCards.first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    // Clicking a hand card toggles its selected state (aria-pressed) and back.
    const firstCard = handCards.first();
    await firstCard.click();
    await expect(firstCard).toHaveAttribute('aria-pressed', 'true', { timeout: TIMEOUT_ACTION });
    await firstCard.click();
    await expect(firstCard).toHaveAttribute('aria-pressed', 'false', { timeout: TIMEOUT_ACTION });

    // The game stays playable (a pass/reset control remains visible).
    await expect(page.getByRole('button', { name: /Pass|パスする|Reset|リセット/ }).first()).toBeVisible({
      timeout: TIMEOUT_ACTION,
    });
  });
});
