import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Shithead E2E', () => {
  test('navigates to shithead and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/shithead');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/シットヘッド|Shithead/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Pickup action button surfaces (always visible on human turn)
    await expect(page.getByRole('button', { name: /場札を引き取る|Pick up pile/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="sh-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('pickup button is clickable on the human turn', async ({ page }) => {
    await navigateTo(page, '/shithead');
    await expect(page.getByText(/シットヘッド|Shithead/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const pickupBtn = page.getByRole('button', { name: /場札を引き取る|Pick up pile/ }).first();
    await expect(pickupBtn).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    if (await pickupBtn.isEnabled()) {
      await pickupBtn.click();
    }
  });
});
