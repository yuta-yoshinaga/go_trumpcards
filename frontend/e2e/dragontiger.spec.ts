import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Dragon Tiger E2E', () => {
  test('plays a round: bet on dragon → auto-resolve → reset', async ({ page }) => {
    await navigateTo(page, '/dragontiger');

    const dragonBtn = page.getByRole('button', { name: 'ドラゴン' });
    await expect(dragonBtn).toBeVisible();
    await dragonBtn.click();
    await waitForLoaded(page);

    // Round resolves immediately — no decision phase. Reset becomes available.
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ドラゴン' })).toBeVisible();
  });
});
