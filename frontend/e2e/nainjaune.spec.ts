import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Le Nain Jaune E2E', () => {
  test('shows the rules, the five boxes and plays a card', async ({ page }) => {
    await navigateTo(page, '/nainjaune');

    // Permanent, not tutorial-only: the run ignoring suit and paying in points
    // are what a player coming from Pope Joan gets wrong.
    await expect(page.getByText(/スート無関係/)).toBeVisible();
    await expect(page.getByText(/枚数ではなく【点数】/)).toBeVisible();

    // All five boxes, always -- the carry-over is only readable there.
    await expect(page.getByTestId('nainjaune-box')).toHaveCount(5, { timeout: TIMEOUT_GAME_LOOP });

    const play = page.getByRole('button', { name: '出す' });
    await expect(play).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    // Nothing is selected yet, so playing is refused.
    await expect(play).toBeDisabled();

    await page.locator('[data-hint-action="play"]').first().click();
    await expect(play).toBeEnabled();
    await play.click();
    await waitForLoaded(page);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/nainjaune');
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
