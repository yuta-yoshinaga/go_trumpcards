import { expect, test } from '@playwright/test';
import { gameButton, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Tien Len E2E', () => {
  test('starts a game: reset → verify controls → pass → reset', async ({ page }) => {
    await navigateTo(page, '/tienlen');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game controls are visible
    const passButton = gameButton(page, 'パス');
    const playButton = page.getByRole('button', { name: '選択したカードを出す' });
    await expect(passButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(playButton).toBeVisible();

    // Pass a turn
    if ((await passButton.getAttribute('aria-disabled')) === 'true') {
      await expect(passButton).toHaveAttribute('aria-disabled', 'true');
    } else {
      await passButton.click();
      await waitForLoaded(page);
    }

    // Game is still running, reset to start fresh
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify controls are back
    await expect(passButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
