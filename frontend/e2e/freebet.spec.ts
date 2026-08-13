import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/** Stands until the round settles, so the assertions below do not depend on the deal. */
async function standOut(page: Parameters<typeof navigateTo>[0]) {
  const stand = page.getByRole('button', { name: 'スタンド' });
  // Splitting produces more hands, each needing its own stand; the bound is a
  // safety net, not an expected count.
  for (let i = 0; i < 8; i++) {
    if (!(await isVisibleWithin(stand, TIMEOUT_ACTION))) return;
    await stand.click();
    await waitForLoaded(page);
  }
}

test.describe('Free Bet Blackjack E2E', () => {
  test('deals, plays a hand out and starts the next round', async ({ page }) => {
    await navigateTo(page, '/freebet');

    const deal = page.getByRole('button', { name: '配る' });
    await expect(deal).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await deal.click();
    await waitForLoaded(page);

    // 伏せ札は無いので、配った時点でディーラーの点数が出る。
    await expect(page.getByTestId('fb-dealer-score')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await standOut(page);

    await expect(page.getByTestId('fb-result')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const next = page.getByRole('button', { name: '次のラウンド' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_ACTION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **無料の操作が出たら、押してもチップが減らないこと。** これがこのゲームの
  // 本体なので、出るまで配り直して 1 回は実際に踏む。
  test('a free raise costs the player nothing', async ({ page }) => {
    await navigateTo(page, '/freebet');

    const chipsOf = async () => {
      const text = (await page.getByTestId('fb-chips').textContent()) ?? '';
      return Number.parseInt(text.replace(/\D/g, ''), 10);
    };

    for (let round = 0; round < 30; round++) {
      await page.getByRole('button', { name: '配る' }).click();
      await waitForLoaded(page);

      const free = page.getByTestId('fb-freedouble').or(page.getByTestId('fb-freesplit')).first();
      if (await isVisibleWithin(free, TIMEOUT_ACTION)) {
        const before = await chipsOf();
        await free.click();
        await waitForLoaded(page);
        // 精算まで進むとチップが動くので、押した直後だけを見る。
        expect(await chipsOf()).toBe(before);
        // ハウス持ちのぶんが別建てで表示される。
        await expect(page.getByTestId('fb-free-0').or(page.getByTestId('fb-free-1')).first()).toBeVisible({
          timeout: TIMEOUT_ACTION,
        });
        return;
      }

      await standOut(page);
      const next = page.getByRole('button', { name: '次のラウンド' });
      if (!(await isVisibleWithin(next, TIMEOUT_ACTION))) break; // out of chips
      await next.click();
      await waitForLoaded(page);
    }
    test.skip(true, '30 ラウンドで無料操作が出る配りに当たらなかった');
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/freebet');
    await expect(page.getByTestId('fb-bet-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

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
