import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Bouillotte E2E', () => {
  // Bouillotte is a 3-card pot-vying game: the human calls, raises (vie), or
  // folds, betting resolves, then the next round can be dealt.
  test('plays a round: bet → resolve → next round', async ({ page }) => {
    await navigateTo(page, '/bouillotte');

    const callButton = page.getByRole('button', { name: /Call|コール/ });
    const foldButton = page.getByRole('button', { name: /Fold|フォールド/ });
    await expect(callButton.or(foldButton).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    await callButton.first().click();
    await waitForLoaded(page);

    const nextRoundButton = page.getByRole('button', { name: /Next Round|次のラウンド/ });
    if (await isVisibleWithin(nextRoundButton.first(), TIMEOUT_ACTION)) {
      await nextRoundButton.first().click();
      await waitForLoaded(page);
      // Back to a betting turn (unless the match ended).
      await expect(callButton.or(page.getByRole('button', { name: /Reset|リセット/ })).first()).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
    }
  });
});
