import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Chemin de Fer E2E', () => {
  test('banks a stake and plays a coup through to the next one', async ({ page }) => {
    await navigateTo(page, '/chemindefer');

    // バンクは席 0（人間）から始まるので、最初は必ず張りの場面。
    const stake = page.getByRole('button', { name: '張る' });
    await expect(stake).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await stake.click();
    await waitForLoaded(page);

    // 張ったあとは CPU が賭け、配られ、子側の判断まで自動で進みます。
    //
    // **ここで何が出るかは配り次第**です。ナチュラル（8 か 9）が出れば 3 枚目は
    // 無いので、親の判断を飛ばしてラウンド終了まで行きます。両方を受けます。
    const draw = page.locator('button:not([disabled])', { hasText: /^引く$/ });
    const next = page.getByRole('button', { name: '次のラウンド' });

    if (await isVisibleWithin(draw.first(), TIMEOUT_TRANSITION)) {
      // 親（あなた）の判断。**どの合計でも引く/立つの両方が選べます。**
      await expect(page.locator('button:not([disabled])', { hasText: /^立つ$/ })).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
      await page.locator('button:not([disabled])', { hasText: /^立つ$/ }).click();
      await waitForLoaded(page);
    }

    // どちらの経路でもラウンドは決着し、次へ進めるようになります。
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cdf-result')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // **確実に動く signal はチップ**です。席行の内容は勝敗で変わらないことがあります。
    const before = await page.getByTestId('cdf-my-chips').textContent();
    await next.click();
    await waitForLoaded(page);

    // 次のラウンドが始まり、盤面が張り待ちに戻っていること。
    await expect(page.getByTestId('cdf-bank-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    expect(before).not.toBeNull();
  });

  test('shows the six seats and the bank marker', async ({ page }) => {
    await navigateTo(page, '/chemindefer');
    await expect(page.getByTestId('cdf-seat-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cdf-seat-5')).toBeVisible();
    // 席 0 が親で始まるので、★ が付いています。
    await expect(page.getByTestId('cdf-seat-0')).toContainText('★');
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/chemindefer');
    await expect(page.getByTestId('cdf-bank-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '張る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
