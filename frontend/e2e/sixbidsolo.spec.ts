import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Six-Bid Solo E2E', () => {
  test('shows the widow and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/sixbidsolo');

    // Permanent, not tutorial-only: eleven cards each and a three-card widow,
    // which is credited to the declarer at the end.
    const widow = page.getByTestId('sixbidsolo-widow');
    await expect(widow).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(widow).toContainText('11枚ずつ配って3枚がウィドウ');

    await expect(page.getByTestId('sixbidsolo-player')).toHaveCount(3, { timeout: TIMEOUT_GAME_LOOP });

    // The human may open the auction, or the CPUs may already have settled it.
    const pass = page.getByRole('button', { name: 'パス' });
    if (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) {
      await pass.click();
      await waitForLoaded(page);
    }

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const declare = page.getByRole('button', { name: '指定する' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(declare, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/sixbidsolo');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('sixbidsolo-widow')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
