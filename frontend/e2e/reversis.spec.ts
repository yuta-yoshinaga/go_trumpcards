import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていない
// （サーバが必ず検証するし、個別に無効化すると「先頭の札を押す」が動く的に
// なる）。その代わり、フォロー義務を満たさない札を押すとサーバが拒否して
// 盤面が動かないので、緑の枠が付いた札を掴む。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Reversis E2E', () => {
  test('navigates to reversis and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/reversis');

    await expect(page.getByText(/レヴェルシ|Reversis/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rv-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **プールと失点配分は盤面から読めない。** 実サーバ相手でも出ていること。
  test('shows the pool and the penalty scale', async ({ page }) => {
    await navigateTo(page, '/reversis');
    // 開始時のプールは全員のアンティ 5×4。
    await expect(page.getByTestId('rv-pool')).toContainText('20', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rv-penalty-rule')).toContainText(/A=4/, { timeout: TIMEOUT_TRANSITION });
  });

  // 4 席すべてが表示され、開始時は全員「無傷」。
  test('shows all four seats, all clean at the start', async ({ page }) => {
    await navigateTo(page, '/reversis');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`rv-seat-${id}`)).toContainText(/無傷|clean/, { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/reversis');
    await expect(page.getByTestId('rv-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('rv-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('rv-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/reversis');
    await expect(page.getByTestId('rv-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('rv-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/reversis');
    await expect(page.getByTestId('rv-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
