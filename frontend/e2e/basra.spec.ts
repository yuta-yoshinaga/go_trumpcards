import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Basra E2E', () => {
  test('loads, resets, and renders the fishing UI', async ({ page }) => {
    await navigateTo(page, '/basra');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    if (await isVisibleWithin(page.getByRole('button', { name: '確認' }), TIMEOUT_TRANSITION)) {
      await page.getByRole('button', { name: '確認' }).click();
    }
    await waitForLoaded(page);

    // Deal / deck info renders.
    await expect(page.getByText(/^配布 \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Some interactive control must be present: the human's play control on their
    // turn, a new-game button once the game ends, or — while a CPU plays — the
    // reset button.
    const anyControl = page
      .getByRole('button', { name: '出す' })
      .or(page.getByRole('button', { name: '捕獲' }))
      .or(page.getByRole('button', { name: '新しいゲーム' }))
      .or(page.getByRole('button', { name: /リセット/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    if (await isVisibleWithin(page.getByRole('button', { name: '確認' }), TIMEOUT_TRANSITION)) {
      await page.getByRole('button', { name: '確認' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^配布 \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
