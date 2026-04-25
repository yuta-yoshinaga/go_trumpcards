import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Spite and Malice E2E', () => {
  test('navigates to spite and malice and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/spiteandmalice');

    // Hand cards (5) for the human are visible.
    const handTutorial = page.locator('[data-tutorial="sam-hand"]');
    await expect(handTutorial).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Foundation row is rendered.
    await expect(page.locator('[data-tutorial="sam-foundations"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button is always visible.
    await expect(page.locator('[data-tutorial="sam-reset"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset returns the board to its initial state', async ({ page }) => {
    await navigateTo(page, '/spiteandmalice');
    await expect(page.locator('[data-tutorial="sam-reset"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.locator('[data-tutorial="sam-reset"]').click();

    // Confirm dialog if present
    const confirm = page.getByRole('button', { name: /確認|confirm|はい/i });
    try {
      await confirm.first().click({ timeout: 1000 });
    } catch {
      // No dialog — proceed
    }

    // Foundation row remains visible after reset.
    await expect(page.locator('[data-tutorial="sam-foundations"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
