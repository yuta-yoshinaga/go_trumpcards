import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

test.describe('tablet navigation', () => {
  test.use({ viewport: { width: 800, height: 900 } });

  test('keeps the game area near the top and provides search in the menu', async ({ page }) => {
    await navigateTo(page, '/hearts');
    const main = await page.locator('main#main-content').boundingBox();
    expect(main?.y).toBeLessThanOrEqual(120);

    await page.getByRole('button', { name: 'メニューを開く' }).click();
    await expect(page.locator('#main-nav').getByPlaceholder('ゲームを検索...')).toBeVisible();
  });

  test('hides the compact menu at large desktop width', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await navigateTo(page, '/hearts');
    await expect(page.getByRole('button', { name: 'メニューを開く' })).toBeHidden();
  });
});
