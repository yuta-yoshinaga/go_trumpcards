import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Baloot E2E', () => {
  test('navigates to baloot and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/baloot');

    await expect(page.getByText(/バルート|Baloot/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('bl-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **宣言は人間まで回ってこないことがある。** 親の左隣から始まるので、手前の
  // CPU が宣言したらその時点でモードが決まる。どちらの場合でも「モードが必ず
  // 決まってプレイに入る」ことを見る。
  const settleMode = async (page: Page) => {
    const sun = page.getByTestId('bl-sun-btn');
    const mode = page.getByTestId('bl-mode');
    // **まず描画を待つ。** いきなり isVisible() を見ると、初回 reset の応答が
    // 届く前に false を返してクリックを飛ばしてしまう。
    await expect(sun.or(mode).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await sun.isVisible()) {
      await sun.click();
    }
    // 決まるとモード行が Sun か Hokom になる。
    await expect(mode).toContainText(/Sun|Hokom/, { timeout: TIMEOUT_ACTION });
    // **人間の手番が来るまで待つ。** リードは親の左隣なので、CPU が先に
    // 打ち終わるまで手札に合法手の枠は付かない。
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles the mode and moves into play', async ({ page }) => {
    await navigateTo(page, '/baloot');
    await expect(page.getByTestId('bl-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await settleMode(page);
    // 決まると宣言ボタンは消える。
    await expect(page.getByTestId('bl-sun-btn')).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });

  // **有効な序列だけが出る。** モードで入れ替わるので、どちらが効いているかを
  // 画面が言い切らないと札の強さが読めない。
  test('states the order that the declared mode puts in force', async ({ page }) => {
    await navigateTo(page, '/baloot');
    await settleMode(page);
    await expect(page.getByTestId('bl-order')).toContainText(/A=11|J=20/, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてがチーム番号つきで出る。
  test('labels all four seats with their team', async ({ page }) => {
    await navigateTo(page, '/baloot');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`bl-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/baloot');
    await settleMode(page);
    await expect(page.getByTestId('bl-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('bl-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('bl-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/baloot');
    await expect(page.getByTestId('bl-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('bl-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/baloot');
    await expect(page.getByTestId('bl-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
