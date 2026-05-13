import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Piquet E2E', () => {
  test('loads page and renders header', async ({ page }) => {
    await navigateTo(page, '/piquet');
    await waitForLoaded(page);
    await expect(page.getByText(/Piquet|ピケ/)).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('reset button restarts the deal', async ({ page }) => {
    await navigateTo(page, '/piquet');
    await waitForLoaded(page);
    const resetButton = page.getByRole('button', { name: /Reset|リセット/ });
    if (await resetButton.isVisible()) {
      await resetButton.click();
      // Some pages open a confirm dialog
      const confirmButton = page.getByRole('button', { name: /Confirm|確認/ });
      if (await confirmButton.isVisible()) await confirmButton.click();
      await waitForLoaded(page);
    }
    await expect(page.getByText(/Piquet|ピケ/)).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
