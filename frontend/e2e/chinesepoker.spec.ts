import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Chinese Poker E2E', () => {
  test('plays a round: bet → set hands → result → reset', async ({ page }) => {
    await navigateTo(page, '/chinesepoker');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // SET HANDS phase: select 3 front + 5 middle cards then click セット
    const cards = page.locator('[data-tutorial="cp-set-hands"] button[aria-label^="Card"]');
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });

    // Click 8 cards to assign front (3) + middle (5)
    for (let i = 0; i < 8; i++) {
      await cards.nth(i).click();
    }

    const setButton = page.getByRole('button', { name: 'セット' });
    await expect(setButton).toBeVisible();
    await setButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
