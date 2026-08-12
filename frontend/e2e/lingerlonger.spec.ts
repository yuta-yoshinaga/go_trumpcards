import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. **The only signal that always moves** when you act. */
const hand = (page: Page) => page.getByRole('button', { name: /を出す$/ });

// **合法な札だけを選ぶ。** フォロー義務を満たさない札はサーバが拒否する。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Linger Longer E2E', () => {
  test('navigates to lingerlonger and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');

    await expect(page.getByText(/リンガーロンガー|Linger Longer/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('ll-stock')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **取っても得点にならず、補充できるだけ。** 直感と逆なので必ず出す。
  test('always states that a trick only buys you a card', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-rule')).toContainText(/補充|stock/, { timeout: TIMEOUT_TRANSITION });
  });

  // **配る枚数は人数と同じ。** 4 人なら 4 枚ずつで、残り 36 枚が山札。
  test('deals four cards each and leaves the rest as stock', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-stock')).toContainText('36', { timeout: TIMEOUT_TRANSITION });
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ll-seat-${id}`)).toContainText('4', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('ll-seat-4')).toHaveCount(0);
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば盤面が動かない。
  test('playing a legal card advances the trick', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const stockBefore = await page.getByTestId('ll-stock').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    // トリック番号か山札のどちらかは必ず動く（取れば補充、取れなければ次の手番）。
    await expect
      .poll(async () => page.getByTestId('ll-stock').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(stockBefore);
  });

  // **勝ち続けるかぎり手札は減らない。** そこがこのゲームの全部。
  test('a seat that wins a trick keeps its hand size', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    for (let i = 0; i < 12; i++) {
      if (await page.getByTestId('ll-result').isVisible()) break;
      const drew = page.getByTestId('ll-seat-0');
      if ((await drew.textContent())?.includes('補充')) {
        // 直前に補充した席は、出した 1 枚が戻っているので配った枚数のまま。
        await expect(drew).toContainText('手札4枚');
        return;
      }
      if (!(await legalCard(page).first().isVisible())) break;
      await legalCard(page).first().click();
      await expect(hand(page).first().or(page.getByTestId('ll-result'))).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
    test.skip(true, 'この配りでは人間が 12 トリック以内に取れなかった');
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-stock')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('ll-stock')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-stock')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('ll-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
