import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Truco E2E', () => {
  test('starts a match: reset → verify game progresses → reset', async ({ page }) => {
    await navigateTo(page, '/truco');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // The match heading and the match-score header should render.
    await expect(page.getByRole('heading', { name: 'トゥルコ' })).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/マッチ得点/)).toBeVisible({ timeout: 10_000 });

    // Reset again to confirm the match can be restarted from any phase.
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'トゥルコ' })).toBeVisible({ timeout: 10_000 });
  });
});
