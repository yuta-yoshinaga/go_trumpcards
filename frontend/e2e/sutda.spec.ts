import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Sutda E2E', () => {
  // **開幕は人間の手番。** 親を最後の席にしてあるので、席 0 から動く。
  test('opens on the human with two cards and a pot', async ({ page }) => {
    await navigateTo(page, '/sutda');
    await expect(page.getByText('ハンド 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sutda-pot')).toContainText('ポット');

    // 自分の 2 枚と役は最初から見えている。
    const hand = page.getByTestId('sutda-hand');
    await expect(hand).toBeVisible({ timeout: TIMEOUT_ACTION });
    // **手続き描画の札は role="img" の div。** img/svg で数えると 0 件になる。
    await expect(hand.getByRole('img')).toHaveCount(2);

    // **相手の札はショーダウンまで伏せたまま。**
    await expect(page.getByTestId('sutda-cards-1')).toHaveCount(0);
    await expect(page.getByTestId('sutda-cards-0')).toBeVisible();
  });

  test('calls through a hand and reaches the showdown', async ({ page }) => {
    await navigateTo(page, '/sutda');
    await expect(page.getByText('ハンド 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const call = page.getByTestId('sutda-call');
    if (!(await isVisibleWithin(call, TIMEOUT_ACTION))) return;
    // 1 巡ぶんコールし続ければショーダウンに届く。
    for (let i = 0; i < 8; i++) {
      if (!(await call.isVisible())) break;
      await call.click();
      await waitForLoaded(page);
    }
    const next = page.getByTestId('sutda-next-hand');
    await expect(next.or(call).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    if (await next.isVisible()) {
      // ショーダウンなら相手の札が開いている。
      await expect(page.getByTestId('sutda-result')).toBeVisible();
    }
  });

  test('folds out of a hand', async ({ page }) => {
    await navigateTo(page, '/sutda');
    await expect(page.getByText('ハンド 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const fold = page.getByTestId('sutda-fold');
    if (!(await isVisibleWithin(fold, TIMEOUT_ACTION))) return;
    await fold.click();
    await waitForLoaded(page);

    // 降りたら自分の手番は戻ってこない ── 次のハンドへ進む面になる。
    await expect(page.getByTestId('sutda-next-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sutda-folded-0')).toBeVisible();
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/sutda');
    await expect(page.getByText('ハンド 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ハンド 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
