import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Joker Poker E2E', () => {
  test('plays a round: deal → draw → result → reset', async ({ page }) => {
    await navigateTo(page, '/jokerpoker');

    // BET phase: click ディール
    const dealButton = page.getByRole('button', { name: 'ディール' });
    await expect(dealButton).toBeVisible();
    await dealButton.click();
    await waitForLoaded(page);

    // DRAW phase: ドロー button should be visible
    const drawButton = page.getByRole('button', { name: 'ドロー' });
    await expect(drawButton).toBeVisible({ timeout: 10_000 });
    await drawButton.click();
    await waitForLoaded(page);

    // RESULT phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase (end state: no confirm dialog)
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ディール' })).toBeVisible();
  });
});
