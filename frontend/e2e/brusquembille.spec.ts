import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Brusquembille E2E', () => {
  test('deals a hand and shows the face-up trump', async ({ page }) => {
    await navigateTo(page, '/brusquembille');
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'ブリュスカンビーユ' })).toBeVisible({
      timeout: 10_000,
    });
    // The face-up trump is what makes the trump suit knowable; without it the
    // page is unplayable.
    await expect(page.getByTestId('brusquembille-stock')).toBeVisible({ timeout: 10_000 });
  });

  test('plays a card while the stock lasts', async ({ page }) => {
    await navigateTo(page, '/brusquembille');
    await waitForLoaded(page);

    // **Assert the hand rendered before asserting anything about a play.**
    // Wrapping the play in `if (visible)` would let this pass on any run that
    // never reached the player's turn.
    const cards = page.getByRole('button', { name: /を出す/ });
    await expect(cards.first(), 'the human hand must render').toBeVisible({ timeout: 10_000 });
    await expect(cards, 'a Brusquembille hand is three cards').toHaveCount(3);

    // While the stock lasts every card is legal — nothing should be disabled.
    for (let i = 0; i < 3; i++) {
      await expect(cards.nth(i), `card ${i} must be playable in the free phase`).toBeEnabled();
    }

    await cards.first().click();
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'ブリュスカンビーユ' })).toBeVisible();
  });
});
