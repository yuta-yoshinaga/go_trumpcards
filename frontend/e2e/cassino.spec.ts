import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Cassino E2E', () => {
  test('navigates to cassino and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/cassino');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/カッシーノ|Cassino/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Action buttons surface
    await expect(page.getByRole('button', { name: /^取る$|^Take$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByRole('button', { name: /^ビルド$|^Build$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByRole('button', { name: /^場に置く$|^Trail$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="cs-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('trail button exists and respects turn state', async ({ page }) => {
    await navigateTo(page, '/cassino');
    await expect(page.getByText(/カッシーノ|Cassino/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const trailBtn = page.getByRole('button', { name: /^場に置く$|^Trail$/ }).first();
    await expect(trailBtn).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Trail requires a hand selection, so the button should be disabled on load.
    await expect(trailBtn).toBeDisabled();
  });
});
