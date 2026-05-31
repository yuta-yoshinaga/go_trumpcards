import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Scopa E2E', () => {
  test('navigates to scopa and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/scopa');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/スコパ|Scopa/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Action buttons surface
    await expect(page.getByRole('button', { name: /^取る$|^Take$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByRole('button', { name: /^場に置く$|^Lay$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="sc-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('lay button exists and respects turn state', async ({ page }) => {
    await navigateTo(page, '/scopa');
    await expect(page.getByText(/スコパ|Scopa/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const layBtn = page.getByRole('button', { name: /^場に置く$|^Lay$/ }).first();
    await expect(layBtn).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Lay requires a hand selection, so the button should be disabled on load.
    await expect(layBtn).toBeDisabled();
  });
});
