import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** ページは出せない札を aria-disabled にするので、
// 手札の先頭を素で掴むと無効な札を押して盤面が動かないことがある。
// `data-legal` は validIndices がそのまま反映された属性。
const playableCard = (page: Page) => page.locator('button[data-legal="true"]');

/** 宣言フェーズを抜ける。契約を選ぶまでプレイは始まらない。 */
async function declare(page: Page, name: RegExp) {
  await page.getByRole('button', { name }).first().click();
}

test.describe('Schafkopf E2E', () => {
  test('navigates to schafkopf and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/schafkopf');

    await expect(page.getByText(/シャーフコップ|Schafkopf/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/ラウンド 1|Round 1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **3 契約すべてに手が届くこと。** Rufspiel だけが押せる状態は、ドメインが
  // 契約を持っているのに誰も宣言できない状態と見分けがつかない。
  test('offers every contract in the declaration phase', async ({ page }) => {
    await navigateTo(page, '/schafkopf');

    await expect(page.getByRole('button', { name: /Rufspiel/ })).toHaveCount(1, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: /Wenz/ })).toHaveCount(1, { timeout: TIMEOUT_TRANSITION });
    // Solo はスートごとに 1 つずつ。宣言してからスートを選ぶ二段構えではない。
    await expect(page.getByRole('button', { name: /Solo/ })).toHaveCount(4, { timeout: TIMEOUT_TRANSITION });
  });

  // **契約は切り札の構成そのもの。** 画面に出ていないと、Wenz の盤面で Ober が
  // 切り札でない理由が読み取れない。
  test('names the contract in play once it is declared', async ({ page }) => {
    await navigateTo(page, '/schafkopf');
    await declare(page, /Wenz/);

    await expect(page.getByText(/契約:.*Wenz|Contract:.*Wenz/).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  // **Wenz と Solo は単独契約。** 呼びフェーズを飛ばしてプレイに入る。
  test('a solo contract skips the call step and starts play', async ({ page }) => {
    await navigateTo(page, '/schafkopf');
    await declare(page, /Wenz/);

    await expect(page.getByRole('button', { name: /Rufspiel/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
    await expect(page.getByRole('button', { name: /を呼ぶ|Call /u })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });

  // Solo は押したボタンのスートがそのまま切り札になる。
  test('a hearts Solo reports hearts as its trump', async ({ page }) => {
    await navigateTo(page, '/schafkopf');
    await declare(page, /♥.*Solo|Solo.*♥/);

    await expect(page.getByText(/契約:.*Solo.*♥|Contract:.*Solo.*♥/).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the trick advances', async ({ page }) => {
    await navigateTo(page, '/schafkopf');
    await declare(page, /Wenz/);
    await expect(playableCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await page
      .getByText(/トリック \d+|Trick \d+/)
      .first()
      .textContent();
    await playableCard(page).first().click();
    await page
      .getByRole('button', { name: /^出す$|^Play$/ })
      .first()
      .click();

    await expect
      .poll(
        async () =>
          page
            .getByText(/トリック \d+|Trick \d+/)
            .first()
            .textContent(),
        { timeout: TIMEOUT_ACTION },
      )
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/schafkopf');
    await expect(page.getByText(/ラウンド 1|Round 1/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /Rufspiel/ })).toHaveCount(1, { timeout: TIMEOUT_ACTION });
  });
});
