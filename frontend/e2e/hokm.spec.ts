import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// 親のときだけ出る切り札ボタン。
const trumpBtn = (page: Page) => page.locator('[data-testid^="hk-trump-"]:not([disabled])');

test.describe('Hokm E2E', () => {
  test('navigates to hokm and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/hokm');

    await expect(page.getByText(/ホクム|Hokm/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('hk-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **13トリック打ち切らないことは盤面から読めない。** 7先取を常に出す。
  test('always shows the race to seven tricks', async ({ page }) => {
    await navigateTo(page, '/hokm');
    await expect(page.getByTestId('hk-race')).toContainText(/7/, { timeout: TIMEOUT_TRANSITION });
  });

  // **親が人間かどうかは配りで決まる。** 親なら宣言し、そうでなければ
  // CPU が宣言済みなので、どちらでもプレイに入ることを見る。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('hk-hand')).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await trumpBtn(page).first().isVisible()) {
      await trumpBtn(page).first().click();
    }
    // **人間の手番が来るまで待つ。** リードは親なので、CPU が先に打つことがある。
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles trump and moves into play', async ({ page }) => {
    await navigateTo(page, '/hokm');
    await settleAndPlay(page);
    await expect(page.getByTestId('hk-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてがチーム番号つきで出る。
  test('labels all four seats with their team', async ({ page }) => {
    await navigateTo(page, '/hokm');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`hk-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick race advances', async ({ page }) => {
    await navigateTo(page, '/hokm');
    await settleAndPlay(page);

    const before = await page.getByTestId('hk-race').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('hk-race').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/hokm');
    await expect(page.getByTestId('hk-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('hk-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/hokm');
    await expect(page.getByTestId('hk-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
