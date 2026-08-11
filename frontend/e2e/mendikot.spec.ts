import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Mendikot E2E', () => {
  test('navigates to mendikot and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/mendikot');

    await expect(page.getByText(/メンディコット|Mendikot/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('md-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **勝敗を決めるのは 10 の枚数。** 盤面から読めないので常に出す。
  test('always shows the race for the four tens', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(page.getByTestId('md-tens')).toContainText(/4/, { timeout: TIMEOUT_TRANSITION });
  });

  // **切り札を選ぶ場面が無い。** 配り終えたらそのままプレイに入る。
  test('starts in play with trump undecided and no trump buttons', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(page.getByTestId('md-trump')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: /^切り札|^Trump / })).toHaveCount(0);
    // 負のコントロール: 押せる札はちゃんと出ている
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてがチーム番号つきで出る。
  test('labels all four seats with their team', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`md-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the board advances', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await page.getByTestId('md-tricks').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('md-tricks').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  // **切り札は打った札で決まる。** 何トリックか進めば必ずスートが立つ。
  test('trump settles once somebody cannot follow', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    for (let i = 0; i < 13; i++) {
      if (/[♠♣♥♦]/.test((await page.getByTestId('md-trump').textContent()) ?? '')) break;
      if (!(await legalCard(page).first().isVisible())) break;
      await legalCard(page).first().click();
      await page.waitForTimeout(150);
    }
    // 決まっていなくてもよい（13 トリック全員フォローできることはある）ので、
    // 決まったなら席の印も一緒に立っていることだけを見る。
    if (/[♠♣♥♦]/.test((await page.getByTestId('md-trump').textContent()) ?? '')) {
      await expect(page.getByText(/切り札決定|set trump/).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(page.getByTestId('md-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('md-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/mendikot');
    await expect(page.getByTestId('md-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
