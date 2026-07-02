import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Trente et Quarante E2E', () => {
  // Trente et Quarante is a pure banking game with no player card decisions:
  // pick a bet, deal, and the round resolves automatically (win/lose/refait).
  test('plays a round: deal → resolve → next round', async ({ page }) => {
    await navigateTo(page, '/trenteetquarante');

    const dealButton = page.getByTestId('teq-deal-button');
    await expect(dealButton).toBeVisible();
    await dealButton.click();
    await waitForLoaded(page);

    const nextRoundButton = page.getByTestId('teq-next-round-button');
    await expect(nextRoundButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await nextRoundButton.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('teq-deal-button')).toBeVisible();
  });
});
