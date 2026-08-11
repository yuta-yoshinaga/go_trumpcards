import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

test.describe('Slobberhannes E2E', () => {
  test('navigates to slobberhannes and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/slobberhannes');

    await expect(page.getByText(/スロバーハンネス|Slobberhannes/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: /^リセット$|^Reset$/ }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  // **最初のトリックは中身に関係なく罰点対象。** 盤面には出ない情報なので、
  // 開始直後に必ず警告が出ていなければならない。
  test('warns that the opening trick carries a penalty', async ({ page }) => {
    await navigateTo(page, '/slobberhannes');
    await expect(page.getByTestId('sh-trick')).toContainText('1/8', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sh-position-warning')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // 4 席すべてが得点と罰の内訳を出す。開始時は全員「無傷」。
  test('shows all four seats, all clean at the start', async ({ page }) => {
    await navigateTo(page, '/slobberhannes');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`sh-seat-${id}`)).toContainText(/無傷|clean/, {
        timeout: TIMEOUT_TRANSITION,
      });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/slobberhannes');
    await expect(page.getByTestId('sh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('sh-trick').textContent();
    const hand = page.getByRole('button', { name: /を出す|^Play / });
    await expect(hand.first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await hand.first().click();

    // 4 人が出し終えるとトリックが解決し、カウンタが進む。
    await expect
      .poll(async () => page.getByTestId('sh-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/slobberhannes');
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
    await navigateTo(page, '/slobberhannes');
    await expect(page.getByTestId('sh-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, {
      timeout: TIMEOUT_ACTION,
    });
  });
});
