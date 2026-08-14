import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// 親のときだけ出る切り札ボタン。
const trumpBtn = (page: Page) => page.locator('[data-testid^="sm-trump-"]:not([disabled])');

test.describe('Sergeant Major E2E', () => {
  test('navigates to sergeantmajor and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');

    await expect(page.getByText(/サージェントメジャー|Sergeant Major/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('sm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **ノルマは席順で決まる。** これが読めないと何をすべきか分からない。
  test('always states that targets follow the seats', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    await expect(page.getByTestId('sm-rule')).toContainText(/席順|seats/, { timeout: TIMEOUT_TRANSITION });
  });

  // **3 席すべてにノルマが出る。**
  test('labels all three seats with their target', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    for (const id of [0, 1, 2]) {
      await expect(page.getByTestId(`sm-seat-${id}`)).toContainText(/ノルマ|target/, { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('sm-seat-3')).toHaveCount(0);
  });

  // **親が人間かどうかは配りで決まる。** どちらでもプレイに入ることを見る。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('sm-round')).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await trumpBtn(page).first().isVisible()) {
      await trumpBtn(page).first().click();
    }
    // 親なら 4 枚選んで捨てる。
    const confirm = page.getByTestId('sm-discard-btn');
    if (await confirm.isVisible()) {
      const cards = page.getByRole('button', { name: /捨て札に選ぶ|Choose / });
      for (let i = 0; i < 4; i++) {
        await cards.nth(i).click();
      }
      await expect(confirm).toBeEnabled({ timeout: TIMEOUT_ACTION });
      await confirm.click();
    }
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles trump and the kitty, then moves into play', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    await settleAndPlay(page);
    await expect(page.getByTestId('sm-trump')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    await settleAndPlay(page);

    // **必ず動く信号は手札の枚数。** 取ったかどうかに依らない。
    const hand = page.getByRole('button', { name: /を出す|^Play / });
    const before = await hand.count();
    expect(before).toBeGreaterThan(0);

    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect.poll(async () => hand.count(), { timeout: TIMEOUT_ACTION }).toBeLessThan(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    await expect(page.getByTestId('sm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('sm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/sergeantmajor');
    await expect(page.getByTestId('sm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('sm-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
