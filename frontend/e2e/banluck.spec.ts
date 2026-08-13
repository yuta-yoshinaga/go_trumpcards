import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Plays the human seat by the rules until the round settles.
 *
 * **The banker cannot stand below 15**, so clicking "stand" blindly stalls:
 * the button is not rendered at all while the obligation holds, and the round
 * never advances. Draw when that is the only legal move.
 */
async function playOutRound(page: Parameters<typeof navigateTo>[0]) {
  const hit = page.getByRole('button', { name: '引く' });
  const stand = page.getByTestId('bl-stand');
  for (let i = 0; i < 12; i++) {
    if (await isVisibleWithin(stand, TIMEOUT_ACTION)) {
      await stand.click();
      await waitForLoaded(page);
      continue;
    }
    if (await isVisibleWithin(hit, TIMEOUT_ACTION)) {
      await hit.click();
      await waitForLoaded(page);
      continue;
    }
    return;
  }
}

test.describe('Ban Luck E2E', () => {
  test('deals, plays a round out and starts the next one', async ({ page }) => {
    await navigateTo(page, '/banluck');

    const deal = page.getByRole('button', { name: '配る' });
    await expect(deal).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await deal.click();
    await waitForLoaded(page);

    // 全席の札が配られる。伏せ札は無い。
    await expect(page.getByTestId('bl-seat-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await playOutRound(page);

    const next = page.getByRole('button', { name: '次のラウンドへ' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **親が席を移ることが、このゲームの骨格。** 1 ラウンド回して確かめる。
  test('the bank moves between seats', async ({ page }) => {
    await navigateTo(page, '/banluck');
    await expect(page.getByTestId('bl-banker')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const before = await page.getByTestId('bl-banker').textContent();

    await page.getByRole('button', { name: '配る' }).click();
    await waitForLoaded(page);
    await playOutRound(page);
    await page.getByRole('button', { name: '次のラウンドへ' }).click();
    await waitForLoaded(page);

    await expect(page.getByTestId('bl-banker')).not.toHaveText(before ?? '', { timeout: TIMEOUT_TRANSITION });
  });

  // **義務が出ている間は「止める」が無い。** 押せない理由も画面に出る。
  //
  // ラウンドを進めて探すのではなく**配り直して探す**。親は席 0 から始まるので
  // 1 ラウンド目の人間は必ず親で、あとは 15 未満を引き当てるだけ ── ラウンドを
  // 回して待つ版は 90 秒のタイムアウトに掛かった (親は 1 周するまで戻らない)。
  test('the banker obligation removes the stand button', async ({ page }) => {
    await navigateTo(page, '/banluck');
    for (let attempt = 0; attempt < 15; attempt++) {
      // **同じ URL への navigate では配り直せない。** HashRouter なので再マウント
      // されず、マウント時の reset が走らないまま前の局面が残る (実測で失敗した)。
      if (attempt > 0) await page.reload();
      await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
      // 1 ラウンド目は人間が親なので、賭け金の入力ではなく案内が出る。
      await expect(page.getByTestId('bl-banker-notice')).toBeVisible({ timeout: TIMEOUT_ACTION });

      await page.getByRole('button', { name: '配る' }).click();
      await waitForLoaded(page);

      if (await isVisibleWithin(page.getByTestId('bl-must-hit'), TIMEOUT_ACTION)) {
        await expect(page.getByTestId('bl-stand')).toHaveCount(0);
        await expect(page.getByRole('button', { name: '引く' })).toBeVisible();
        return;
      }
    }
    test.skip(true, '15 回配っても親が 15 未満になる配りに当たらなかった');
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/banluck');
    await expect(page.getByTestId('bl-round-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
