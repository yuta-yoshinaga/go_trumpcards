import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Nertz / Pounce E2E', () => {
  test('navigates to nertz and renders the shared foundation row', async ({ page }) => {
    await navigateTo(page, '/nertz');

    // Page title visible (Japanese or English label).
    await expect(page.getByText(/ナーツ|Nertz/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button is reachable via the shared data-tutorial hook.
    await expect(page.locator('[data-tutorial="nertz-reset"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    // The shared foundation grid renders the first foundation cell (F0).
    // aria-label uses the localized template, so accept either ja or en.
    await expect(page.getByLabel(/ファウンデーション0|Foundation 0/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('renders the human player tableau and stock controls', async ({ page }) => {
    await navigateTo(page, '/nertz');
    await expect(page.getByText(/ナーツ|Nertz/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Tutorial-tagged regions must surface for the human player area.
    await expect(page.locator('[data-tutorial="nertz-pile"]').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="nertz-tableau"]').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="nertz-stock"]').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
