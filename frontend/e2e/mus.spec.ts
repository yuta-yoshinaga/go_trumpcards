import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Mus E2E', () => {
  test('loads, resets, and renders interactive phase UI', async ({ page }) => {
    await navigateTo(page, '/mus');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round info renders.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Some interactive control must be present: Mus/Corte, a bet action, the
    // next-round button, or the reset button once a CPU-driven phase resolves.
    const corteButton = page.getByRole('button', { name: /コルテ（勝負）/ });
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });
    await expect
      .poll(
        async () => (await page.getByTestId('game-footer-actions').isVisible()) || (await anyResetButton.isVisible()),
        { timeout: TIMEOUT_GAME_LOOP },
      )
      .toBe(true);

    // Exercise one stable interaction when the human is asked to cut (Corte),
    // which deterministically starts the betting rounds without touching cards.
    if (await isVisibleWithin(corteButton, TIMEOUT_TRANSITION)) {
      await corteButton.click();
      await waitForLoaded(page);
    }

    // The board still shows round info (the game advanced, not crashed).
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
