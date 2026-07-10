import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Michigan E2E', () => {
  // Michigan is a "stops" chip-betting game: the human first distributes chips
  // across the four boodles, then plays cards in ascending same-suit runs.
  test('places boodle bets and enters the play phase', async ({ page }) => {
    await navigateTo(page, '/michigan');

    // Bet phase: the place-bets button is preloaded with an even distribution.
    const placeBets = page.getByRole('button', { name: /Place bets|賭ける/ });
    await expect(placeBets.first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    await placeBets.first().click();
    await waitForLoaded(page);

    // After betting, the round proceeds — either the human gets a play turn
    // (playable hand cards / next round) or the round resolves to a reset point.
    const handCard = page.getByTestId('hand-card-0');
    const nextRound = page.getByRole('button', { name: /Next Round|次のラウンド/ });
    const resetButton = page.getByRole('button', { name: /Reset|リセット/ });
    await expect(handCard.or(nextRound).or(resetButton).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    if (await isVisibleWithin(nextRound.first(), TIMEOUT_ACTION)) {
      await nextRound.first().click();
      await waitForLoaded(page);
    }
  });
});
