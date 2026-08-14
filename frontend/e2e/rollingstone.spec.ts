import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. **The only signal that always moves** when you act. */
const hand = (page: Page) => page.getByRole('button', { name: /を出す$/ });

// **合法な札だけを選ぶ。** フォロー義務を満たさない札はサーバが拒否する。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Rolling Stone E2E', () => {
  test('navigates to rollingstone and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/rollingstone');

    await expect(page.getByText(/ローリングストーン|Rolling Stone/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('rs-deck')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **勝利条件が逆さまなのが規則そのもの。**
  test('always states that taking tricks is worth nothing', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(page.getByTestId('rs-rule')).toContainText(/得点|score/, { timeout: TIMEOUT_TRANSITION });
  });

  // **4 人なら 32 枚、1 人 8 枚。**
  test('deals eight cards to each of the four seats', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(page.getByTestId('rs-deck')).toContainText('32', { timeout: TIMEOUT_TRANSITION });
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`rs-seat-${id}`)).toContainText('8', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('rs-seat-4')).toHaveCount(0);
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば手札が動かない。
  test('acting removes a card from your hand, or hands you the trick', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await hand(page).count();
    expect(before).toBeGreaterThan(0);

    // 出せる札があれば出す。無ければ引き取りだけが出ているはず。
    const pickup = page.getByTestId('rs-pickup-btn');
    if (await pickup.isVisible()) {
      await pickup.click();
      // **引き取りは手札が増える。** 罰則がそのまま盤面に出る。
      await expect.poll(async () => hand(page).count(), { timeout: TIMEOUT_ACTION }).toBeGreaterThan(before);
      return;
    }

    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();
    await expect.poll(async () => hand(page).count(), { timeout: TIMEOUT_ACTION }).not.toBe(before);
  });

  // **出せる札が無いときは、手札を押させない。**
  test('disables the hand whenever a pickup is forced', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    // 引き取りが要る局面に当たるまで進める。
    for (let i = 0; i < 20; i++) {
      if (await page.getByTestId('rs-pickup-btn').isVisible()) {
        await expect(page.getByTestId('rs-must-pickup')).toBeVisible({ timeout: TIMEOUT_ACTION });
        await expect(hand(page).first()).toBeDisabled();
        return;
      }
      if (!(await legalCard(page).first().isVisible())) break;
      await legalCard(page).first().click();
      await expect(hand(page).first().or(page.getByTestId('rs-result'))).toBeVisible({ timeout: TIMEOUT_ACTION });
      if (await page.getByTestId('rs-result').isVisible()) break;
    }
    test.skip(true, 'この配りでは引き取りが起きなかった');
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(page.getByTestId('rs-deck')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('rs-deck')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/rollingstone');
    await expect(page.getByTestId('rs-deck')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('rs-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
