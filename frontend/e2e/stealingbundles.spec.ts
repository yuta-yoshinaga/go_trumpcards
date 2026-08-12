import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. Selecting a card is the first half of every move. */
const hand = (page: Page) => page.getByRole('button', { name: /を選ぶ$/ });

test.describe('Stealing Bundles E2E', () => {
  test('navigates to stealingbundles and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');

    await expect(page.getByText(/スティーリングバンドル|Stealing Bundles/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('sb-table')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **束の一番上が弱点、というのが規則そのもの。**
  test('always states that a bundle can be taken whole', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(page.getByTestId('sb-rule')).toContainText(/束|bundle/, { timeout: TIMEOUT_TRANSITION });
  });

  // **場に4枚、各自に4枚。** 4 人なら山札は 32 枚。
  test('deals four to the table and four to each of the four seats', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(page.getByTestId('sb-header')).toContainText('32', { timeout: TIMEOUT_TRANSITION });
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`sb-seat-${id}`)).toContainText('4', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('sb-seat-4')).toHaveCount(0);
  });

  // **選ぶだけでは盤面は動かない。** 手を決めるのは 2 手目のボタンです。
  test('selecting a card offers its legal moves and only then acts', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('sb-actions')).toHaveCount(0);

    const before = await page.getByTestId('sb-header').textContent();
    // 何かしら手のある札を探す（緑の枠が付いている札）。
    const usable = page.locator('button.ring-ds-success');
    await expect(usable.first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await usable.first().click();
    await expect(page.getByTestId('sb-actions')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // 出ているボタンのどれかを押すと盤面が進む。
    const act = page
      .getByTestId('sb-take-btn')
      .or(page.getByTestId('sb-trail-btn'))
      .or(page.locator('[data-testid^="sb-steal-btn-"]'));
    await act.first().click();
    await expect
      .poll(async () => page.getByTestId('sb-header').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  // **取れるときは置けません。** 促しとボタンの両方が一致していること。
  test('never offers trailing while a capture is available', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(page.getByTestId('sb-status')).toBeVisible({ timeout: TIMEOUT_ACTION });

    const status = await page.getByTestId('sb-status').textContent();
    const mustCapture = /場に置くことはできません|cannot place a card/.test(status ?? '');
    await hand(page).first().click();
    await expect(page.getByTestId('sb-actions')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('sb-trail-btn')).toHaveCount(mustCapture ? 0 : 1);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(page.getByTestId('sb-table')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('sb-table')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/stealingbundles');
    await expect(page.getByTestId('sb-table')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('sb-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
