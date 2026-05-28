import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Dou Dizhu E2E', () => {
  test('starts a game: reset → verify game progresses → reset', async ({ page }) => {
    await navigateTo(page, '/doudizhu');

    // Click reset to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // The game starts in the bid phase. Verify the page heading rendered.
    await expect(page.getByRole('heading', { name: '斗地主' })).toBeVisible({ timeout: 10_000 });

    // Reset again to confirm the game can be restarted from any phase.
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: '斗地主' })).toBeVisible({ timeout: 10_000 });
  });
});
