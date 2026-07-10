import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Guts E2E', () => {
  // Guts is a fast pot-vying gamble: the human declares In (stay) or Out (fold),
  // the round resolves, then the next round can be dealt.
  test('plays a round: declare → resolve → next round', async ({ page }) => {
    await navigateTo(page, '/guts');

    const inButton = page.getByRole('button', { name: /In \(stay\)|イン/ });
    const outButton = page.getByRole('button', { name: /Out \(fold\)|アウト/ });
    await expect(inButton.or(outButton).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    await inButton.first().click();
    await waitForLoaded(page);

    const nextRoundButton = page.getByRole('button', { name: /Next Round|次のラウンド/ });
    if (await isVisibleWithin(nextRoundButton.first(), TIMEOUT_ACTION)) {
      await nextRoundButton.first().click();
      await waitForLoaded(page);
      // Back to a declare turn (unless the match ended).
      await expect(inButton.or(page.getByRole('button', { name: /Reset|リセット/ })).first()).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
    }
  });
});
