import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Agnes Sorel E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/agnes');

    await expect(page.getByText(/ベースランク/)).toBeVisible();
    await expect(page.getByText(/手数/).first()).toBeVisible();
  });

  test('deal advances the game', async ({ page }) => {
    await navigateTo(page, '/agnes');

    const dealButton = page.getByRole('button', { name: '配る' });
    await expect(dealButton).toBeVisible();
    await dealButton.click();
    await waitForLoaded(page);
  });

  test('reset button restarts the game', async ({ page }) => {
    await navigateTo(page, '/agnes');

    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);
  });
});
