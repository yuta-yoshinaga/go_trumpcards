import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// 人間の番のときだけ出る宣言ボタン。
const bidBtn = (page: Page) => page.locator('[data-testid^="hpf-bid-"]:not([disabled])');
const passBtn = (page: Page) => page.getByTestId('hpf-pass-btn');
const discardBtn = (page: Page) => page.locator('[data-testid^="hpf-discard-"]');

test.describe('Hasenpfeffer E2E', () => {
  test('navigates to hasenpfeffer and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');

    await expect(page.getByText(/ハーゼンプフェファー|Hasenpfeffer/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **ジョーカーが最強という序列は知らないと打ち方が変わる。**
  test('always states the joker ranking', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await expect(page.getByTestId('hpf-rule')).toContainText(/Best Bower/, { timeout: TIMEOUT_TRANSITION });
  });

  test('labels all four seats with their team', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`hpf-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  // **競り → 捨て札 → プレイの 3 段を抜ける。** どこで人間の番が来るかは配り次第。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // 競り: 落札を狙う（降りると次のハンドまで手番が来ないことがある）。
    if (await bidBtn(page).first().isVisible()) {
      await bidBtn(page).last().click();
    }
    // 捨て札: 落札していれば札を選んでスートを決める。
    if (await discardBtn(page).first().isVisible()) {
      await page
        .getByRole('button', { name: /捨て札に選ぶ|Choose / })
        .first()
        .click();
      await discardBtn(page).first().click();
    }
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles the auction and moves into play', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await settleAndPlay(page);
    await expect(page.getByTestId('hpf-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await settleAndPlay(page);

    // **必ず動く信号は手札の枚数。** 取ったかどうかに依らない。
    const hand = page.getByRole('button', { name: /を出す|^Play / });
    const before = await hand.count();
    expect(before).toBeGreaterThan(0);

    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect.poll(async () => hand.count(), { timeout: TIMEOUT_ACTION }).toBeLessThan(before);
  });

  // **降りられない場面では降りるボタンを出さない。** 出ているなら押せること。
  test('offers a pass only when passing is allowed', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    if (await page.getByTestId('hpf-must-bid').isVisible()) {
      await expect(passBtn(page)).toHaveCount(0);
      await expect(bidBtn(page).first()).toBeVisible();
    } else if (await passBtn(page).isVisible()) {
      await expect(passBtn(page)).toBeEnabled();
    }
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/hasenpfeffer');
    await expect(page.getByTestId('hpf-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('hpf-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
