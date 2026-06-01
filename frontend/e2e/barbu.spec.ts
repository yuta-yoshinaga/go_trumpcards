import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Barbu E2E', () => {
  test('navigates to barbu and renders the contract-selection phase', async ({ page }) => {
    await navigateTo(page, '/barbu');

    // Page title visible (either Japanese or English label).
    await expect(page.getByText(/バルブ|Barbu/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // The human deals first, so contract buttons should surface.
    await expect(page.getByTestId('contract-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('contract-6')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Reset button via data-tutorial to avoid i18n coupling.
    await expect(page.locator('[data-tutorial="bb-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('selecting a contract advances into the play phase', async ({ page }) => {
    await navigateTo(page, '/barbu');
    await expect(page.getByTestId('contract-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Pick "No Tricks" — a non-trump contract that starts trick play immediately.
    await page.getByTestId('contract-0').click();

    // Play button surfaces once the deal moves into the play phase.
    await expect(page.getByTestId('play-button')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
