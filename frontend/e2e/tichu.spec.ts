import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Tichu E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/tichu');

    // CPU player areas always render once the game state loads.
    await expect(page.getByText(/CPU 1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset button is present', async ({ page }) => {
    await navigateTo(page, '/tichu');

    await expect(page.getByText(/CPU 1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="tichu-reset"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('table area renders', async ({ page }) => {
    await navigateTo(page, '/tichu');

    await expect(page.getByText(/CPU 1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.locator('[data-tutorial="tichu-table"]').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
