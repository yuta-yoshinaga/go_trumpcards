import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Blackjack Switch E2E', () => {
  test('plays a round: bet → keep → stand both → reset', async ({ page }) => {
    await navigateTo(page, '/blackjackswitch');

    // BET phase: place a bet (per-hand amount; total cost = 2x).
    const betButton = page.getByRole('button', { name: 'ベットする' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // SWITCH phase appears unless the dealer is dealt a natural BJ (rare).
    // We pick "Keep" to keep the test deterministic regardless of which cards
    // were dealt.
    const keepButton = page.getByRole('button', { name: 'そのまま' });
    if (await keepButton.isVisible()) {
      await keepButton.click();
      await waitForLoaded(page);
    }

    // ACTION phase: stand on each hand until the round resolves. Loop a few
    // times to advance through both hands without depending on shuffle order.
    for (let i = 0; i < 4; i++) {
      const standButton = page.getByRole('button', { name: 'スタンド' });
      if (!(await standButton.isVisible())) break;
      await standButton.click();
      await waitForLoaded(page);
    }

    // Round resolves; reset (next-game) button appears.
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベットする' })).toBeVisible();
  });
});
