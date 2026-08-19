import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('BlackJack E2E', () => {
  test('plays a round: bet → stand → result', async ({ page }) => {
    await navigateTo(page, '/');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION or INSURANCE phase should appear after bet
    // If insurance is offered, decline it
    const declineButton = page.getByRole('button', { name: '辞退' });
    if (await isVisibleWithin(declineButton, TIMEOUT_ACTION)) {
      await declineButton.click();
      await waitForLoaded(page);
    }

    // A natural blackjack (dealer's, or every player hand) settles the round in
    // checkNaturalBlackJack without ever entering the ACTION phase, so スタンド
    // never appears on those deals -- wait for either state rather than assuming
    // one (#6063).
    const standButton = page.getByRole('button', { name: 'スタンド' });
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(standButton.or(resetButton).first()).toBeVisible({ timeout: 5_000 });

    if (await standButton.isVisible()) {
      await standButton.click();
      await waitForLoaded(page);
    }

    // END phase: 次のゲーム button should be visible either way
    await expect(resetButton).toBeVisible({ timeout: 10_000 });
  });
});
