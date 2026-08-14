import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Extra Bet Blackjack E2E', () => {
  test('deals, declines the extra bet and plays a hand out', async ({ page }) => {
    await navigateTo(page, '/doubleattack');

    const deal = page.getByRole('button', { name: '配る' });
    await expect(deal).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await deal.click();
    await waitForLoaded(page);

    // **追加ベットの前は 1 枚だけで、点数が出ない。**
    await expect(page.getByTestId('da-dealer-hidden')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('da-dealer-score')).toHaveCount(0);
    await expect(page.getByTestId('da-attack-notice')).toBeVisible();

    await page.getByRole('button', { name: '見送る' }).click();
    await waitForLoaded(page);

    // 2 枚目が配られたら点数が出る。
    await expect(page.getByTestId('da-dealer-score')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('da-dealer-hidden')).toHaveCount(0);

    // ブラックジャックだと即決着するので、プレイ操作は出ているときだけ。
    const stand = page.getByRole('button', { name: 'スタンド' });
    if (await isVisibleWithin(stand, TIMEOUT_ACTION)) {
      await stand.click();
      await waitForLoaded(page);
    }

    await expect(page.getByTestId('da-result')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const next = page.getByRole('button', { name: '次のラウンド' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_ACTION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('raising is offered up to the ante', async ({ page }) => {
    await navigateTo(page, '/doubleattack');
    await page.getByRole('button', { name: '配る' }).click();
    await waitForLoaded(page);

    // **上限はサーバの値。** ラベルに出ているので、賭け増しの導線があることだけ見る。
    await expect(page.getByRole('button', { name: '賭け増す' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '見送る' })).toBeVisible();

    await page.getByRole('button', { name: '賭け増す' }).click();
    await waitForLoaded(page);
    await expect(page.getByTestId('da-dealer-score')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/doubleattack');
    await expect(page.getByTestId('da-bet-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

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
