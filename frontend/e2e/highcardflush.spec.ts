import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('High Card Flush E2E', () => {
  test('plays a round: bet → raise (1x always available) → result → reset', async ({ page }) => {
    await navigateTo(page, '/highcardflush');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION phase: 1x is always allowed regardless of flush length
    const raise1x = page.getByRole('button', { name: 'レイズ x1' });
    await expect(raise1x).toBeVisible({ timeout: 10_000 });
    await raise1x.click();
    await waitForLoaded(page);

    // END phase: reset button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('fold flow: bet → fold → result → reset', async ({ page }) => {
    await navigateTo(page, '/highcardflush');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    const foldButton = page.getByRole('button', { name: 'フォールド' });
    await expect(foldButton).toBeVisible({ timeout: 10_000 });
    await foldButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
