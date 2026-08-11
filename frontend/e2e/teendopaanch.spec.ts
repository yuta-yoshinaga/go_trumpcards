import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// ノルマ 5 の席のときだけ出る切り札ボタン。
const trumpBtn = (page: Page) => page.locator('[data-testid^="td-trump-"]:not([disabled])');

test.describe('3-2-5 E2E', () => {
  test('navigates to teendopaanch and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');

    await expect(page.getByText(/3-2-5/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('td-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **ノルマは宣言ではなく割り当て。** これが読めないと何をすべきか分からない。
  test('always states that the targets are assigned', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    await expect(page.getByTestId('td-targets')).toContainText(/割り当て|assigned/, { timeout: TIMEOUT_TRANSITION });
  });

  // **3 席すべてにノルマが出る。**
  test('labels all three seats with their target', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    for (const id of [0, 1, 2]) {
      await expect(page.getByTestId(`td-seat-${id}`)).toContainText(/ノルマ|target/, { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('td-seat-3')).toHaveCount(0);
  });

  // **ノルマ 5 が人間かどうかは配りで決まる。** どちらでもプレイに入ることを見る。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('td-round')).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await trumpBtn(page).first().isVisible()) {
      await trumpBtn(page).first().click();
    }
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles trump and moves into play', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    await settleAndPlay(page);
    await expect(page.getByTestId('td-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  // **席行は変わらないことがある。** ノルマ・獲得数・達成回数しか出ておらず、
  // 打った札がトリックを取らなければ 3 つとも同じまま。**手札が 1 枚減ったこと**を
  // 見るのが、勝敗に依らず必ず動く唯一の信号。
  test('plays a card and the board advances', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    await settleAndPlay(page);

    const hand = page.getByRole('button', { name: /を出す|^Play / });
    const before = await hand.count();
    expect(before).toBeGreaterThan(0);

    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect.poll(async () => hand.count(), { timeout: TIMEOUT_ACTION }).toBeLessThan(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    await expect(page.getByTestId('td-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('td-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/teendopaanch');
    await expect(page.getByTestId('td-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('td-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
