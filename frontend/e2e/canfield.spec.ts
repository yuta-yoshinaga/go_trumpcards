import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Canfield E2E', () => {
  test('navigates and renders the game', async ({ page }) => {
    await navigateTo(page, '/canfield');

    // **本文で拾わない。** ベースランクの循環規則を盤に出した (#7188) ので、
    // `/ベースランク/` は見出しと規則文の 2 つに当たり strict mode で落ちる。
    await expect(page.getByTestId('cf-base-rank')).toBeVisible();
    await expect(page.getByTestId('cf-base-rank-rule')).toBeVisible();
    await expect(page.getByText(/手数/).first()).toBeVisible();
  });

  test('draw from stock advances move count', async ({ page }) => {
    await navigateTo(page, '/canfield');

    const stockButton = page.getByRole('button', { name: /山札/ });
    await expect(stockButton).toBeVisible();
    await stockButton.click();
    await waitForLoaded(page);
  });

  test('reset button restarts the game', async ({ page }) => {
    await navigateTo(page, '/canfield');

    const resetButton = page.getByRole('button', { name: 'リセット' }).first();
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);
  });
});
