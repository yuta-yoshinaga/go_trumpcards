import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Desmoche E2E', () => {
  test('shows both rules and draws a card', async ({ page }) => {
    await navigateTo(page, '/desmoche');

    // Permanent, not tutorial-only: going out takes ten rather than the nine
    // dealt, and poker rankings play no part.
    await expect(page.getByText(/ちょうど10枚/)).toBeVisible();
    await expect(page.getByText(/ポーカーの役は使いません/)).toBeVisible();
    await expect(page.getByText(/ポット/)).toBeVisible();

    // The human draws first, so the draw controls must be there.
    const drawStock = page.getByRole('button', { name: '山札から引く' });
    await expect(drawStock).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await drawStock.click();
    await waitForLoaded(page);

    // After drawing, the act controls appear.
    await expect(page.getByRole('button', { name: 'メルド' })).toBeVisible();
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/desmoche');

    await page.getByRole('button', { name: '山札から引く' }).click();
    await waitForLoaded(page);

    // Five rounds run before the game ends, so the reset control may read
    // either label.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/ラウンド: 1/)).toBeVisible();
  });
});
