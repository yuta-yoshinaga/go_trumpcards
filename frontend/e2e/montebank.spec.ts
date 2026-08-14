import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Monte Bank E2E', () => {
  test('bets, turns the gate and starts the next round', async ({ page }) => {
    await navigateTo(page, '/montebank');

    // 場札 4 枚が並ぶ。
    await expect(page.getByTestId('mb-layout-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('mb-layout-3')).toBeVisible();
    // 賭ける前にゲートは出ない。
    await expect(page.getByTestId('mb-gate')).toHaveCount(0);

    await page.getByRole('button', { name: '賭ける' }).click();
    await waitForLoaded(page);

    await expect(page.getByTestId('mb-gate')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('mb-result')).toBeVisible();

    const next = page.getByRole('button', { name: '次のラウンドへ' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_ACTION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '賭ける' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **どの札を選んだかが、そのままサーバに届く。** 0 始まりの位置を送るので、
  // ここがずれると「左端を選んだのに別の札で判定される」形で静かに壊れる。
  test('the picked card is the one that settles', async ({ page }) => {
    await navigateTo(page, '/montebank');
    await expect(page.getByTestId('mb-layout-2')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByTestId('mb-layout-2').click();
    await expect(page.getByTestId('mb-layout-2')).toHaveAttribute('aria-pressed', 'true');

    await page.getByRole('button', { name: '賭ける' }).click();
    await waitForLoaded(page);

    // 決着後は選んだ札に印が残る (サーバが pick を返している)。
    await expect(page.getByTestId('mb-layout-2')).toHaveClass(/ring-ds-success/, { timeout: TIMEOUT_TRANSITION });
  });

  // **各札に「互角 / 不利」が必ず付く。** それが賭けの良し悪しを決める唯一の
  // 数字なので、出ていなければ勝負が運任せになる。
  test('every layout card is labelled even or against', async ({ page }) => {
    await navigateTo(page, '/montebank');
    await expect(page.getByTestId('mb-note-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    for (let i = 0; i < 4; i++) {
      await expect(page.getByTestId(`mb-note-${i}`)).toHaveText(/互角|不利/);
    }
  });

  // **山を使い切ると終わる。** 40 枚 / 5 枚 = 8 ラウンド。
  test('plays through the deck and finishes', async ({ page }) => {
    await navigateTo(page, '/montebank');

    for (let round = 0; round < 12; round++) {
      const bet = page.getByRole('button', { name: '賭ける' });
      if (!(await isVisibleWithin(bet, TIMEOUT_ACTION))) break;
      await bet.click();
      await waitForLoaded(page);

      const next = page.getByRole('button', { name: '次のラウンドへ' });
      if (!(await isVisibleWithin(next, TIMEOUT_ACTION))) break;
      await next.click();
      await waitForLoaded(page);
    }

    // 8 ラウンド後は賭けも次も出ない。
    await expect(page.getByRole('button', { name: '賭ける' })).toHaveCount(0, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '次のラウンドへ' })).toHaveCount(0);
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/montebank');
    await expect(page.getByTestId('mb-round-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '賭ける' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
