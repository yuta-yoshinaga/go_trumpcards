import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Bhabhi E2E', () => {
  test('navigates to bhabhi and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/bhabhi');

    await expect(page.getByText(/バービー|Bhabhi/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **勝者ではなく敗者を決めるゲーム。** 目的が読めないと何をすべきか分からない。
  test('always states the goal of not being the Bhabhi', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(page.getByTestId('bh-goal')).toContainText(/Bhabhi/, { timeout: TIMEOUT_TRANSITION });
  });

  // **ハンドの区切りが無い。** 次のハンドへ進むボタンは出ない。
  test('offers no next-hand button', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: /^次の|^Next/ })).toHaveCount(0);
    // 負のコントロール: リセットは出ている
    await expect(page.getByRole('button', { name: /^リセット$|^Reset$/ }).first()).toBeVisible();
  });

  test('renders every seat at the default table', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`bh-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the board advances', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await page.getByTestId('bh-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('bh-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  // **人数を変えるとゲームの形が変わる。** 席がその数だけ出ること。
  //
  // 人数の select は SettingsPanel の中で、**既定で閉じた `<details>`** なので
  // 先に summary を開かないと操作できない（jsdom は閉じた details の中身も
  // 拾ってしまうので、ここでしか気付けない）。
  test('deals a bigger table when the player count changes', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await settingsToggle.click();

    const playerCnt = page.getByTestId('bh-player-cnt');
    await expect(playerCnt).toBeVisible({ timeout: TIMEOUT_ACTION });
    await playerCnt.selectOption('6');

    await expect(page.getByTestId('bh-seat-5')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('bh-seat-6')).toHaveCount(0);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/bhabhi');
    await expect(page.getByTestId('bh-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('bh-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
