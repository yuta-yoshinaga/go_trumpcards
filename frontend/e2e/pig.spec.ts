import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. **The only signal that always moves** when you act. */
const hand = (page: Page) => page.getByRole('button', { name: /へ渡す$/ });

test.describe('Pig E2E', () => {
  test('navigates to pig and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/pig');

    await expect(page.getByText(/ピッグ|Pig/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('pig-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **取り合うものが何もないのが規則そのもの。**
  test('always states that the signal is silent', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(page.getByTestId('pig-rule')).toContainText(/鼻|nose/, { timeout: TIMEOUT_TRANSITION });
  });

  // **デッキは人数 × 4 枚。** 4 人なら 16 枚で、全員が 4 枚ずつ。
  test('deals four cards to each of the four seats', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(page.getByTestId('pig-round')).toContainText('16', { timeout: TIMEOUT_TRANSITION });
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`pig-seat-${id}`)).toContainText('4', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('pig-seat-4')).toHaveCount(0);
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば盤面が動かない。
  test('passing a card advances the round', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await page.getByTestId('pig-round').textContent();
    await expect(hand(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await hand(page).first().click();

    // パス回数が増えるか、合図／ラウンド終了へ進むかのどちらかは必ず起きる。
    await expect
      .poll(async () => page.getByTestId('pig-round').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  // **合図が出たら押すボタンは1つだけ。** 遅れることだけが負け。
  test('offers only the signal button once a signal is out', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    for (let i = 0; i < 30; i++) {
      if (await page.getByTestId('pig-signal-btn').isVisible()) {
        await expect(page.getByTestId('pig-signal-alert')).toBeVisible();
        await expect(hand(page).first()).toBeDisabled();
        await page.getByTestId('pig-signal-btn').click();
        // 名乗ったあとは合図のボタンが消える。
        await expect(page.getByTestId('pig-signal-btn')).toHaveCount(0, { timeout: TIMEOUT_ACTION });
        return;
      }
      if (await page.getByTestId('pig-next-btn').isVisible()) {
        await page.getByTestId('pig-next-btn').click();
        await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
        continue;
      }
      if (await page.getByTestId('pig-result').isVisible()) break;
      if (!(await hand(page).first().isEnabled())) break;
      await hand(page).first().click();
      await expect(hand(page).first().or(page.getByTestId('pig-result'))).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
    test.skip(true, 'この配りでは合図が人間の番に回ってこなかった');
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(page.getByTestId('pig-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('pig-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/pig');
    await expect(page.getByTestId('pig-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('pig-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
