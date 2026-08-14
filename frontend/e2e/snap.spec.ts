import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

test.describe('Snap E2E', () => {
  test('navigates to snap and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/snap');

    await expect(page.getByText(/スナップ|Snap/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sp-pile')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **トリガーが動くことが規則そのもの。**
  test('always states the rule', async ({ page }) => {
    await navigateTo(page, '/snap');
    await expect(page.getByTestId('sp-rule')).toContainText(/ランク|rank/, { timeout: TIMEOUT_TRANSITION });
  });

  test('shows both stocks, 26 each at the start', async ({ page }) => {
    await navigateTo(page, '/snap');
    await expect(page.getByTestId('sp-seat-0')).toContainText('26', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sp-seat-1')).toContainText('26', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('sp-seat-2')).toHaveCount(0);
  });

  // **必ず動く信号はストックの枚数。** 場札は取られると 0 に戻る。
  test('turning a card moves it out of your stock', async ({ page }) => {
    await navigateTo(page, '/snap');
    const step = page.getByTestId('sp-step-btn');
    await expect(step).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await step.click();

    await expect
      .poll(async () => (await page.getByTestId('sp-seat-0').textContent()) ?? '', { timeout: TIMEOUT_ACTION })
      .not.toContain('26');
  });

  // **宣言はいつでも押せる。** 成立していなければペナルティになるだけ。
  test('the snap button is always pressable, and a wrong call costs a card', async ({ page }) => {
    await navigateTo(page, '/snap');
    const snap = page.getByTestId('sp-snap-btn');
    await expect(snap).toBeEnabled({ timeout: TIMEOUT_ACTION });

    // 開始直後は場札が空なので、宣言すれば必ず誤宣言になる。
    await snap.click();
    await expect(page.getByTestId('sp-event')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect
      .poll(async () => (await page.getByTestId('sp-seat-0').textContent()) ?? '', { timeout: TIMEOUT_ACTION })
      .not.toContain('26');
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/snap');
    await expect(page.getByTestId('sp-pile')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('sp-pile')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/snap');
    await expect(page.getByTestId('sp-pile')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('sp-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
