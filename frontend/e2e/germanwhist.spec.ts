import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていない
// （サーバが必ず検証するし、個別に無効化すると「先頭の札を押す」が動く的に
// なる）。その代わり、フォロー義務を満たさない札を押すとサーバが拒否して
// 盤面が動かないので、緑の枠が付いた札を掴む。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('German Whist E2E', () => {
  test('navigates to germanwhist and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/germanwhist');

    await expect(page.getByText(/ジャーマンホイスト|German Whist/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('gw-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: /^リセット$|^Reset$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  // 前半は表向きの札を奪い合うためのトリック。その 1 枚が出ていなければ
  // ゲームとして成立しないので、盤面ではなくこれを見る。
  test('shows the face-up card the first half is played for', async ({ page }) => {
    await navigateTo(page, '/germanwhist');
    await expect(page.getByTestId('gw-upcard')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('gw-upcard')).toContainText(/表向きの札|Face-up card/);
  });

  // 実際にカードを出して、サーバが param error ではなく次の局面を返すこと。
  //
  // **手札の枚数は進行の指標にならない。** 前半は 1 トリックごとに勝者が
  // 表向きの札を、敗者が山札の札を引くので、手札は 13 枚のまま動かない。
  // これはバグではなくこのゲームの定義そのものなので、下で明示的に固定する。
  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/germanwhist');
    await expect(page.getByTestId('gw-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('gw-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('gw-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('the hand stays at 13 cards through the first half', async ({ page }) => {
    await navigateTo(page, '/germanwhist');
    const hand = page.getByRole('button', { name: /を出す|^Play / });
    await expect(hand).toHaveCount(13, { timeout: TIMEOUT_TRANSITION });

    await legalCard(page).first().click();
    // トリックが解決すると双方が 1 枚ずつ補充するので、また 13 枚に戻る。
    await expect(hand).toHaveCount(13, { timeout: TIMEOUT_ACTION });
    // 減ったのは山札のほう。
    await expect(page.getByTestId('gw-upcard')).toBeVisible();
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/germanwhist');
    await expect(page.getByTestId('gw-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('gw-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/germanwhist');
    await expect(page.getByTestId('gw-phase')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    // 終局すると投了ボタンが消える。
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, {
      timeout: TIMEOUT_ACTION,
    });
  });
});
