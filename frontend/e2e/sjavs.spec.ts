import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Sjavs E2E', () => {
  test('shows the permanent trumps and bids', async ({ page }) => {
    await navigateTo(page, '/sjavs');

    // Permanent, not tutorial-only: believing only the trump suit is trump is
    // the standard mistake in this game.
    await expect(page.getByText(/♣Q ＞ ♠Q ＞ ♣J ＞ ♠J ＞ ♥J ＞ ♦J/)).toBeVisible();
    // Trump is only decided by bidding, so it must not name a suit yet.
    await expect(page.getByText(/切札: 未定/)).toBeVisible();

    // Bidding is the first thing the human does. Pass is always available; a
    // length button only exists when the deal supports it, so pass is what the
    // test can rely on across shuffles.
    const pass = page.getByRole('button', { name: 'パス' });
    await expect(pass).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await pass.click();
    await waitForLoaded(page);

    // After passing, either the hand is under way or a new deal is being bid.
    // Both keep the rule line on screen, which is what this asserts.
    await expect(page.getByText(/♣Q ＞ ♠Q/)).toBeVisible();
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/sjavs');

    await page.getByRole('button', { name: 'パス' }).click();
    await waitForLoaded(page);

    // The rubber runs over many hands, so the reset control may read either
    // label depending on how far the CPUs got. Match both.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    // JSX puts whitespace between the expressions, so the rendered text is
    // "味方 24" rather than "味方24".
    await expect(page.getByText(/残り: 味方\s*24/)).toBeVisible();
  });
});
