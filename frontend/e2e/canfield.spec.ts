import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Canfield E2E', () => {
  test('keeps selected tableau move controls visible on large desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await navigateTo(page, '/canfield');
    const card = page.locator('[data-tutorial="cf-tableau"] button').first();
    await expect(card).toBeVisible();
    await card.click();
    const bar = page.getByTestId('cf-selected-move-bar');
    await expect(bar).toBeVisible();
    const bounds = await bar.boundingBox();
    if (!bounds) throw new Error('Canfield move bar has no bounding box');
    expect(bounds.y).toBeGreaterThanOrEqual(0);
    expect(bounds.y + bounds.height).toBeLessThanOrEqual(800);
    await expect(bar.getByRole('button').first()).toBeVisible();
  });

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
