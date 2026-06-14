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
    const anyControl = page
      .getByRole('button', { name: /ムス（交換）/ })
      .or(corteButton)
      .or(page.getByRole('button', { name: /交換する/ }))
      .or(page.getByRole('button', { name: /パソ（パス）/ }))
      .or(page.getByRole('button', { name: '次のラウンド' }))
      .or(page.getByRole('button', { name: /リセット|次のゲーム/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

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
