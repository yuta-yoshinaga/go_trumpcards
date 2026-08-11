import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

// 押せる入札だけを選ぶ。上回れない額のボタンはそもそも描かれない。
const openBid = (page: Page) => page.locator('[data-testid^="sh-bid-"]:not([disabled])');

test.describe('Shelem E2E', () => {
  test('navigates to shelem and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/shelem');

    await expect(page.getByText(/シェレム|Shelem/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **点になるのは A/10/5 だけ。** 盤面から読めないので常時出ていること。
  test('always states the point cards', async ({ page }) => {
    await navigateTo(page, '/shelem');
    await expect(page.getByTestId('sh-points')).toContainText(/100/, { timeout: TIMEOUT_TRANSITION });
  });

  // **競り → ウィドウ整理 → プレイ の3段。** 人間が落札するかどうかで
  // 通る道が変わるので、どちらでもプレイに入ることを見る。
  const settleAndPlay = async (page: Page) => {
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // 1 段階目: 降りられるなら降りる。降りられない席では入札する。
    const pass = page.getByTestId('sh-pass-btn');
    if (await pass.isVisible()) {
      await pass.click();
    } else if (await openBid(page).first().isVisible()) {
      await openBid(page).first().click();
    }

    // 2 段階目: 自分が落札したら 4 枚選んで切り札を決める。
    const discard = page.getByTestId('sh-discard-1-btn');
    if (await discard.isVisible()) {
      const picks = page.getByRole('button', { name: /捨て札に選ぶ|Select .* to discard/ });
      for (let i = 0; i < 4; i++) {
        await picks.nth(i).click();
      }
      await expect(discard).toBeEnabled({ timeout: TIMEOUT_ACTION });
      await discard.click();
    }

    // **人間の手番が来るまで待つ。** リードは落札者なので CPU が先に打つ。
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  test('settles the contract and moves into play', async ({ page }) => {
    await navigateTo(page, '/shelem');
    await settleAndPlay(page);
    await expect(page.getByTestId('sh-contract')).toContainText(/[♠♣♥♦]/, { timeout: TIMEOUT_ACTION });
  });

  // 4 席すべてが立場つきで出る。
  test('labels all four seats', async ({ page }) => {
    await navigateTo(page, '/shelem');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`sh-seat-${id}`)).toContainText(/T[01]/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/shelem');
    await settleAndPlay(page);
    await expect(page.getByTestId('sh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('sh-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('sh-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/shelem');
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/shelem');
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
