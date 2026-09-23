import { expect, test } from '@playwright/test';
import { gameButton, navigateTo, waitForLoaded } from './helpers';

test.describe('Baccarat E2E', () => {
  test('plays a round: bet → result → reset', async ({ page }) => {
    await navigateTo(page, '/baccarat');

    // BET phase: click ベット
    const betButton = gameButton(page, 'ベット');
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });
});
