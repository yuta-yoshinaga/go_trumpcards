import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Escoba E2E', () => {
  test('navigates to escoba and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/escoba');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/エスコバ|Escoba/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Per-player score surfaces (i18n-agnostic via test id)
    await expect(page.getByTestId('player-score-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // A hand card or an action button should be present
    await expect(
      page
        .getByTestId('hand-card-0')
        .or(page.getByRole('button', { name: /^取る$|^Take$/ }))
        .first(),
    ).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="es-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('reset restarts the game', async ({ page }) => {
    await navigateTo(page, '/escoba');
    await expect(page.getByText(/エスコバ|Escoba/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset (clicking the reset control, then confirming if a dialog appears)
    await page.locator('[data-tutorial="es-reset-button"]').first().click({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click({ timeout: TIMEOUT_TRANSITION })
      .catch(() => {});

    // Game state remains rendered after reset
    await expect(page.getByTestId('player-score-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
