import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// **押せる宣言だけを選ぶ。** ノルマ未満と合計13になる値は disabled なので、
// 決め打ちの番号を押すとその値に当たったときだけ落ちる。
const openBid = (page: Page) => page.locator('[data-testid^="iw-bid-"]:not([disabled])');

test.describe('Israeli Whist E2E', () => {
  test('navigates to israeliwhist and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');

    await expect(page.getByText(/イスラエリホイスト|Israeli Whist/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('iw-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **的中が2乗で跳ねることは盤面から読めない。** 常時出ていること。
  test('always states the scoring', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    await expect(page.getByTestId('iw-score')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('iw-score')).not.toBeEmpty();
  });

  // **入札が2段階あるぶん、人間まで回ってこない段もある。** 出ているボタンを
  // 押し、どちらの段から入ってもプレイに到達することを見る。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('iw-round')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // 1 段階目: 出ていれば降りる（降りられない席なら入札に切り替わる）。
    const pass = page.getByTestId('iw-pass-btn');
    if (await pass.isVisible()) {
      await pass.click();
    }

    // 2 段階目: 押せる宣言があれば押す。
    await expect(openBid(page).first().or(legalCard(page).first())).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await openBid(page).first().isVisible()) {
      await openBid(page).first().click();
    }

    // **人間の手番が来るまで待つ。** リードは落札者なので、CPU が先に打つ。
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles both bidding rounds and moves into play', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    await settleAndPlay(page);
    await expect(page.getByTestId('iw-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてが立場と宣言つきで出る。
  test('labels all four seats', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`iw-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    await settleAndPlay(page);
    await expect(page.getByTestId('iw-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('iw-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('iw-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    await expect(page.getByTestId('iw-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('iw-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/israeliwhist');
    await expect(page.getByTestId('iw-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
