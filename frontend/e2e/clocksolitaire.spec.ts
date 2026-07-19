import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_LOADED, waitForLoaded } from './helpers';

test.describe('Clock Solitaire E2E', () => {
  test('navigates and plays steps', async ({ page }) => {
    await navigateTo(page, '/clocksolitaire');

    // Verify step count is displayed
    await expect(page.getByText(/ステップ数/)).toBeVisible();

    // Click step button
    const stepButton = page.getByRole('button', { name: 'ステップ' });
    await expect(stepButton).toBeVisible();
    await stepButton.click();
    await waitForLoaded(page);
  });

  test('autoplay completes the game', async ({ page }) => {
    // Autoplay now animates each move via a client-side stepped loop; seed the
    // fastest speed so the full run stays well within the wait budget.
    await page.addInitScript(() => {
      window.localStorage.setItem('clocksolitaire:autoPlaySpeed', 'fast');
    });
    await navigateTo(page, '/clocksolitaire');

    // Click autoplay
    const autoplayButton = page.getByRole('button', { name: 'オートプレイ' });
    await expect(autoplayButton).toBeVisible();
    await autoplayButton.click();

    // The stepped autoplay animates every placement until the game ends, at
    // which point the step/autoplay controls are removed. Wait for completion.
    await expect(page.getByRole('button', { name: 'ステップ' })).not.toBeVisible({
      timeout: TIMEOUT_LOADED,
    });
  });
});
