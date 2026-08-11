import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていない
// （サーバが必ず検証するし、個別に無効化すると「先頭の札を押す」が動く的に
// なる）。その代わり、フォロー義務を満たさない札を押すとサーバが拒否して
// 盤面が動かないので、緑の枠が付いた札を掴む。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Polignac E2E', () => {
  test('navigates to polignac and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/polignac');

    await expect(page.getByText(/ポリニャック|Polignac/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('pg-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **失点の重み付けは盤面から読み取れない。** 常時出ていなければならない。
  test('always states the jack penalty rule', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await expect(page.getByTestId('pg-penalty-rule')).toContainText(/Polignac/, { timeout: TIMEOUT_TRANSITION });
  });

  // **配り直後は宣言フェーズ。** ここを逃すと capot は打てないので、
  // 実際に両方のボタンが出ていることを本物のサーバで確かめる。
  test('opens on the declaration phase with capot and pass', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await expect(page.getByTestId('pg-capot-btn')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('pg-pass-btn')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('passing the declaration starts play and hides the buttons', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await page.getByTestId('pg-pass-btn').click();

    await expect(page.getByTestId('pg-capot-btn')).toHaveCount(0, { timeout: TIMEOUT_ACTION });
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
  });

  // capot を宣言すると、宣言バナーが実際に出る。
  test('declaring capot announces it', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await page.getByTestId('pg-capot-btn').click();

    await expect(page.getByTestId('pg-capot-banner')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('pg-seat-0')).toContainText('[capot]', { timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await page.getByTestId('pg-pass-btn').click();
    await expect(page.getByTestId('pg-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('pg-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('pg-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await expect(page.getByTestId('pg-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('pg-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/polignac');
    await expect(page.getByTestId('pg-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
