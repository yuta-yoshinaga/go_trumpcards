import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

test.describe('Coinche E2E', () => {
  test('navigates to coinche and renders the opening board', async ({ page }) => {
    await navigateTo(page, '/coinche');

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText('チームスコア').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **競りは 4 スートすべてに開いている。** ベロートのような表向き札の除外は
  // 無いので、開幕は 12 契約 × 4 スートから選べる。
  test('opens the auction with every suit available', async ({ page }) => {
    await navigateTo(page, '/coinche');

    await expect(page.getByLabel('目標点')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    for (const suit of ['♠ スペード', '♣ クラブ', '♥ ハート', '♦ ダイヤ']) {
      await expect(page.getByRole('button', { name: `${suit}で宣言する` })).toHaveCount(1, {
        timeout: TIMEOUT_TRANSITION,
      });
    }
    await expect(page.getByRole('button', { name: 'パス' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **契約は「点 + 切り札」の対。** 点を選ばないうちにスートだけ押せると、
  // 残りに既定値が入って別の契約になる。
  test('a suit cannot be bid before a target is chosen', async ({ page }) => {
    await navigateTo(page, '/coinche');

    const spade = page.getByRole('button', { name: '♠ スペードで宣言する' });
    await expect(spade).toBeDisabled({ timeout: TIMEOUT_TRANSITION });

    await page.getByLabel('目標点').selectOption('250');
    await expect(spade).toBeEnabled({ timeout: TIMEOUT_ACTION });
  });

  // Capot (250) は最上位契約なので、宣言すれば必ず競りに勝つ。
  test('a Capot bid wins the auction and is named on the board', async ({ page }) => {
    await navigateTo(page, '/coinche');

    await page.getByLabel('目標点').selectOption('250');
    await page.getByRole('button', { name: '♠ スペードで宣言する' }).click();

    await expect(page.getByTestId('co-contract')).toContainText('250', { timeout: TIMEOUT_ACTION });
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/coinche');
    await expect(page.getByLabel('目標点')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByLabel('目標点')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
