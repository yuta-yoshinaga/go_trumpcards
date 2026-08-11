import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Tarabish E2E', () => {
  test('navigates to tarabish and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/tarabish');

    await expect(page.getByText(/タラビッシュ|Tarabish/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('tb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **切り札の序列はこの系統の肝。** 盤面から読めないので常時出ていること。
  test('always states the trump order', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    await expect(page.getByTestId('tb-order')).toContainText(/Jass/, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('tb-order')).toContainText(/Menel/, { timeout: TIMEOUT_TRANSITION });
  });

  // **入札は人間まで回ってこないことがある。** 親の左隣から始まるので、
  // 手前の CPU が引き受けたらその時点で切り札が決まる。どちらの場合でも
  // 「切り札が必ず決まってプレイに入る」ことを見る。
  const settleTrump = async (page: Page) => {
    const take = page.getByTestId('tb-take-btn');
    const trump = page.getByTestId('tb-trump');
    // **まず描画を待つ。** いきなり isVisible() を見ると、初回 reset の応答が
    // 届く前に false を返してクリックを飛ばしてしまう。
    await expect(take.or(trump).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await take.isVisible()) {
      await take.click();
    }
    // 決まったら切り札の表示に変わる。
    await expect(trump).toBeVisible({ timeout: TIMEOUT_ACTION });
    // **人間の手番が来るまで待つ。** リードは親の左隣なので、CPU が先に
    // 打ち終わるまで手札に合法手の枠は付かない。
    await expect(page.locator('button.ring-ds-success').first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles trump and moves into play', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    await expect(page.getByTestId('tb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await settleTrump(page);
    // 切り札が決まると候補の表示は消える。
    await expect(page.getByTestId('tb-upcard')).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてがチーム番号つきで出る。
  test('labels all four seats with their team', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`tb-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    await settleTrump(page);
    await expect(page.getByTestId('tb-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('tb-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('tb-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    await expect(page.getByTestId('tb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('tb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/tarabish');
    await expect(page.getByTestId('tb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
