import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Poch E2E', () => {
  test('shows the rules, the nine pools and the betting controls', async ({ page }) => {
    await navigateTo(page, '/poch');

    // Permanent, not tutorial-only: paying only on the turn-up's suit and
    // pochen being a comparison rather than a declaration are the two rules a
    // player gets wrong.
    await expect(page.getByText(/めくり札と同じスート/)).toBeVisible();
    await expect(page.getByText(/宣言ではなく同ランクの組の比べ合い/)).toBeVisible();

    // All nine pools, always -- the carry-over is only readable from them.
    await expect(page.getByTestId('poch-pool')).toHaveCount(9, { timeout: TIMEOUT_GAME_LOOP });

    // The human acts first in the pochen.
    const bet = page.getByRole('button', { name: '賭ける' });
    await expect(bet).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await bet.click();
    await waitForLoaded(page);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/poch');
    await waitForLoaded(page);

    // Five deals run before the game ends, so the reset control may read
    // either label.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/ディール1\/5/)).toBeVisible();
  });
});
