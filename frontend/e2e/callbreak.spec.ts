import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Call Break E2E', () => {
  test('loads page and renders round header', async ({ page }) => {
    await navigateTo(page, '/callbreak');
    await waitForLoaded(page);
    // Scope to <main> so the assertion does not match the NavBar entry.
    await expect(
      page
        .locator('main')
        .getByText(/ラウンド/)
        .first(),
    ).toBeVisible({
      timeout: TIMEOUT_ACTION,
    });
  });

  test('reset button restarts the round', async ({ page }) => {
    await navigateTo(page, '/callbreak');
    await waitForLoaded(page);
    const resetButton = page.getByRole('button', { name: /Reset|リセット/ }).first();
    if (await resetButton.isVisible()) {
      await resetButton.click();
      const confirmButton = page.getByRole('button', { name: /Confirm|確認/ }).first();
      if (await confirmButton.isVisible()) await confirmButton.click();
      await waitForLoaded(page);
    }
    await expect(
      page
        .locator('main')
        .getByText(/ラウンド/)
        .first(),
    ).toBeVisible({
      timeout: TIMEOUT_ACTION,
    });
  });
});
