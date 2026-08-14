import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// **押せる宣言だけを選ぶ。** 合計が13になる 1 つだけが disabled になるので、
// 決め打ちの番号を押すとその値に当たったときだけ落ちる。
const openBid = (page: Page) => page.locator('[data-testid^="est-bid-"]:not([disabled])');

test.describe('Estimation E2E', () => {
  test('navigates to estimation and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/estimation');

    await expect(page.getByText(/エスティメーション|Estimation/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('est-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **的中だけが得点になることは盤面から読めない。** 常時出ていること。
  test('always states the scoring', async ({ page }) => {
    await navigateTo(page, '/estimation');
    await expect(page.getByTestId('est-score')).toContainText(/Dash Call/, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('est-score')).toContainText(/Risk/, { timeout: TIMEOUT_TRANSITION });
  });

  // **切り札選択と宣言は人間まで回ってこないことがある。** 親が CPU なら
  // 切り札はもう決まっているし、宣言も順番次第。どちらでも「プレイに入る」
  // ことを見る。
  const settleAndPlay = async (page: Page) => {
    const trump = page.getByTestId('est-trump-1-btn');
    // **まず描画を待つ。** いきなり isVisible() を見ると、初回 reset の応答が
    // 届く前に false を返してクリックを飛ばしてしまう。
    await expect(page.getByTestId('est-round')).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await trump.isVisible()) {
      await trump.click();
    }
    await expect(openBid(page).first().or(legalCard(page).first())).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await openBid(page).first().isVisible()) {
      await openBid(page).first().click();
    }
    // **人間の手番が来るまで待つ。** リードは親の左隣なので、CPU が先に
    // 打ち終わるまで手札に合法手の枠は付かない。
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles trump and calls, then moves into play', async ({ page }) => {
    await navigateTo(page, '/estimation');
    await settleAndPlay(page);
    await expect(page.getByTestId('est-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてが宣言と累計つきで出る。
  test('labels all four seats', async ({ page }) => {
    await navigateTo(page, '/estimation');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`est-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/estimation');
    await settleAndPlay(page);
    await expect(page.getByTestId('est-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('est-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('est-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/estimation');
    await expect(page.getByTestId('est-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('est-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/estimation');
    await expect(page.getByTestId('est-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
