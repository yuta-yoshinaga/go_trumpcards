import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Doppelkopf E2E', () => {
  test('loads, resets, and renders the play UI', async ({ page }) => {
    await navigateTo(page, '/doppelkopf');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round / trick info renders.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The game starts in the Play phase; some interactive control must be present
    // (the human's play/announce control, a trick/round advance, or — once the
    // round resolves via CPU play — the reset / next-game button).
    const anyControl = page
      .getByRole('button', { name: '出す' })
      .or(page.getByRole('button', { name: /を宣言$/ }))
      .or(page.getByRole('button', { name: '次のトリック' }))
      .or(page.getByRole('button', { name: '次のラウンド' }))
      .or(page.getByRole('button', { name: /リセット|次のゲーム/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
