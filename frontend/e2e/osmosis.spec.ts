import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Osmosis E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/osmosis');

    await expect(page.getByText(/ベースランク/)).toBeVisible();
    await expect(page.getByText(/手数/).first()).toBeVisible();
  });

  test('draw from stock advances the game', async ({ page }) => {
    await navigateTo(page, '/osmosis');

    const stockButton = page.getByRole('button', { name: /山札/ });
    await expect(stockButton).toBeVisible();
    await stockButton.click();
    await waitForLoaded(page);
  });

  test('reset button restarts the game', async ({ page }) => {
    await navigateTo(page, '/osmosis');

    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);
  });
});
