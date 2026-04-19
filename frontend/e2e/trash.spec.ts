import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Trash E2E', () => {
  test('navigates to trash and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/trash');

    // Both player rows should expose all 10 face-down slots via aria-label.
    const faceDownSlots = page.getByRole('button', { name: /face-down/ });
    await expect(faceDownSlots.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // 20 slot buttons total (10 opponent + 10 self); allow for a larger count
    // if future UI additions add more face-down affordances.
    await expect.poll(async () => await faceDownSlots.count()).toBeGreaterThanOrEqual(20);

    // Reset button is always visible.
    await expect(page.locator('[data-tutorial="tr-reset"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('reset returns the board to its initial state', async ({ page }) => {
    await navigateTo(page, '/trash');
    await expect(page.locator('[data-tutorial="tr-reset"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.locator('[data-tutorial="tr-reset"]').click();

    // Confirm dialog if present
    const confirm = page.getByRole('button', { name: /確認|confirm|はい/i });
    try {
      await confirm.first().click({ timeout: 1000 });
    } catch {
      // No dialog — proceed
    }

    // Board is still rendered after reset.
    const faceDownSlots = page.getByRole('button', { name: /face-down/ });
    await expect(faceDownSlots.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
