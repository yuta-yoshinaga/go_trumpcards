import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Pope Joan E2E', () => {
  test('shows the rules, the eight compartments and plays a card', async ({ page }) => {
    await navigateTo(page, '/popejoan');

    // Permanent, not tutorial-only: compartments pay only on trumps, and the
    // missing eight of diamonds is what makes a run always die at the seven.
    await expect(page.getByText(/トランプの札でしか取れません/)).toBeVisible();
    await expect(page.getByText(/♦8 が抜いてあるので/)).toBeVisible();

    // All eight compartments, always -- the carry-over is only readable there.
    await expect(page.getByTestId('popejoan-compartment')).toHaveCount(8, { timeout: TIMEOUT_GAME_LOOP });

    const play = page.getByRole('button', { name: '出す' });
    await expect(play).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    // Nothing is selected yet, so playing is refused.
    await expect(play).toBeDisabled();

    // **出せない札は disabled。**先頭の札が常に選べるとは限らない (#4933-#4935)。
    await page.locator('[data-hint-action="play"]:not([disabled])').first().click();
    await expect(play).toBeEnabled();
    await play.click();
    await waitForLoaded(page);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/popejoan');
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
