import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Schnapsen E2E', () => {
  test('navigates to schnapsen and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/schnapsen');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/シュナプセン|Schnapsen/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Phase indicator surfaces (phase 1 on a fresh deal)
    await expect(page.getByTestId('schnapsen-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Reset button surfaces
    await expect(page.getByRole('button', { name: /^リセット$|^Reset$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/schnapsen');
    await expect(page.getByText(/シュナプセン|Schnapsen/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    // Confirm dialog appears
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    // Game still renders after reset
    await expect(page.getByTestId('schnapsen-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
