import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Sheepshead E2E', () => {
  test('loads, resets, and renders interactive phase UI', async ({ page }) => {
    await navigateTo(page, '/sheepshead');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round / trick info renders.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The game starts in the Pick phase; some interactive control must be present
    // (pick/pass for the human, or — when a CPU is the picker — a later-phase
    // control or the reset button once the round resolves via CPU play).
    const pickButton = page.getByRole('button', { name: 'ピックする' });
    const passButton = page.getByRole('button', { name: 'パスする' });
    const anyControl = pickButton
      .or(passButton)
      .or(page.getByRole('button', { name: '出す' }))
      .or(page.getByRole('button', { name: '次のトリック' }))
      .or(page.getByRole('button', { name: '次のラウンド' }))
      .or(page.getByRole('button', { name: /リセット|次のゲーム/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Exercise one stable interaction when the human is asked to pick/pass.
    if (await isVisibleWithin(passButton, TIMEOUT_TRANSITION)) {
      await passButton.click();
      await waitForLoaded(page);
    } else if (await isVisibleWithin(pickButton, TIMEOUT_TRANSITION)) {
      await pickButton.click();
      await waitForLoaded(page);
    }

    // The board still shows round info (the game advanced, not crashed).
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
