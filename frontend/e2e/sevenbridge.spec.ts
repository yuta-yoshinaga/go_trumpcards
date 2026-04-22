import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Seven Bridge E2E', () => {
  test('navigates to sevenbridge and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/sevenbridge');

    // Round indicator visible
    await expect(page.getByText(/ラウンド|Round/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Draw phase action button
    await expect(page.getByRole('button', { name: /山札から引く|Draw from stock/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="sb-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('drawing from stock advances the phase', async ({ page }) => {
    await navigateTo(page, '/sevenbridge');

    await expect(page.getByText(/ラウンド|Round/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /山札から引く|Draw from stock/ })
      .first()
      .click();

    // Meld or discard controls should appear during the play phase
    const meldBtn = page.getByRole('button', { name: /メルド|Meld/ });
    const discardBtn = page.getByRole('button', { name: /捨てる|Discard/ });
    await expect(meldBtn.or(discardBtn).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
