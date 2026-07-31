import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Vint E2E', () => {
  test('shows the reversed ranking and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/vint');

    // Permanent, not tutorial-only: spades are the LOWEST denomination, the
    // reverse of bridge, and there is no dummy.
    const ladder = page.getByTestId('vint-ladder');
    await expect(ladder).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(ladder).toContainText('♠');
    await expect(ladder).toContainText('NT');
    await expect(page.getByTestId('vint-no-dummy')).toContainText('ダミーはありません');

    await expect(page.getByTestId('vint-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('vint-scores')).toBeVisible();

    // The human may open the auction, or the CPUs may already have settled it.
    const pass = page.getByRole('button', { name: 'パス' });
    if (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) {
      await pass.click();
      await waitForLoaded(page);
    }

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/vint');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('vint-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
