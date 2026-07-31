import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Loba E2E', () => {
  test('shows both rules and draws a card', async ({ page }) => {
    await navigateTo(page, '/loba');

    // Permanent, not tutorial-only: "three different suits" and the joker
    // restriction are what a player gets wrong.
    await expect(page.getByText(/異なる3スート/)).toBeVisible();
    await expect(page.getByText(/ジョーカーは1枚まで、ピエルナ不可/)).toBeVisible();
    await expect(page.getByText(/101点で脱落/)).toBeVisible();

    // The human draws first, so the draw controls must be there.
    const drawStock = page.getByRole('button', { name: '山札から引く' });
    await expect(drawStock).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await drawStock.click();
    await waitForLoaded(page);

    // After drawing, the act controls appear.
    await expect(page.getByRole('button', { name: 'メルド' })).toBeVisible();
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/loba');

    await page.getByRole('button', { name: '山札から引く' }).click();
    await waitForLoaded(page);

    // Many rounds run before anyone is knocked out, so the reset control may
    // read either label.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/ラウンド: 1/)).toBeVisible();
  });
});
