import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Deuces Wild E2E', () => {
  test('plays a round: deal → draw → result → reset', async ({ page }) => {
    await navigateTo(page, '/deuceswild');

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

    // RESULT phase: 次のハンド button should be visible
    const resetButton = page.getByRole('button', { name: '次のハンド' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ディール' })).toBeVisible();
  });
});
