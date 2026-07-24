import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Zheng Shangyou E2E', () => {
  test('starts a game: reset → verify controls → toggle card selection', async ({ page }) => {
    await navigateTo(page, '/zheng');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game controls are visible. Pass may be legitimately disabled when
    // the human leads (empty table), so only assert visibility here.
    const passButton = page.getByRole('button', { name: 'パス' });
    const playButton = page.getByRole('button', { name: '選択したカードを出す' });
    await expect(passButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(playButton).toBeVisible();
    await expect(playButton).toBeDisabled(); // nothing selected yet

    // After the CPU replay it is the human's turn: selecting a hand card
    // enables Play (a single is always a legal shape), deselecting disables it.
    const firstCard = page.getByTestId('hand-card-0');
    await firstCard.click({ timeout: TIMEOUT_GAME_LOOP });
    await expect(playButton).toBeEnabled({ timeout: TIMEOUT_GAME_LOOP });
    await firstCard.click();
    await expect(playButton).toBeDisabled();

    // Reset works again from a running game
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(playButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
