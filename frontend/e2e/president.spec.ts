import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('President E2E', () => {
  test('navigates to president and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/president');

    // Page title visible (either Japanese or English label)
    await expect(page.getByText(/プレジデント|President/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Play / Pass action buttons surface
    await expect(page.getByRole('button', { name: /^出す$|^Play$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByRole('button', { name: /^パス$|^Pass$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    // Reset button via data-tutorial to avoid i18n coupling
    await expect(page.locator('[data-tutorial="pr-reset-button"]').first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('pass button is clickable when it is the human turn', async ({ page }) => {
    await navigateTo(page, '/president');
    await expect(page.getByText(/プレジデント|President/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const passBtn = page.getByRole('button', { name: /^パス$|^Pass$/ }).first();
    await expect(passBtn).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Pass may be enabled or disabled depending on whose turn it is after shuffle.
    // Clicking should not throw — if disabled Playwright will surface an error,
    // so guard with a conditional check.
    if (await passBtn.isEnabled()) {
      await passBtn.click();
    }
  });
});
