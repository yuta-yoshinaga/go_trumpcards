import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your bid cards. */
const hand = (page: Page) => page.getByRole('button', { name: /で入札する$/ });

test.describe('Goofspiel E2E', () => {
  test('navigates to goofspiel and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/goofspiel');

    await expect(page.getByText(/ゴフスピール|Goofspiel/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('gs-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **同時入札であることが規則そのもの。**
  test('always states that bids are simultaneous', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(page.getByTestId('gs-rule')).toContainText(/同時|same time/, { timeout: TIMEOUT_TRANSITION });
  });

  // **賞札と懸かっている点は常に出ている。**
  test('shows the prize on the table', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(page.getByTestId('gs-prize')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **13 枚配って 12 枚残る。** 1 枚めくられた状態で始まります。
  test('deals thirteen bid cards and turns the first prize', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(hand(page)).toHaveCount(13, { timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('gs-header')).toContainText('12', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('gs-seat-2')).toHaveCount(0);
  });

  // **入札すると同時に公開される。** 1 手で結果まで進みます。
  test('bidding resolves the round in one step', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    await hand(page).first().click();
    await expect(page.getByTestId('gs-round-end')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('gs-next-btn')).toBeVisible();
    // 使った札は戻らない。
    await expect(hand(page)).toHaveCount(12, { timeout: TIMEOUT_ACTION });

    await page.getByTestId('gs-next-btn').click();
    await expect(page.getByTestId('gs-prize')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(page.getByTestId('gs-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('gs-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/goofspiel');
    await expect(page.getByTestId('gs-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('gs-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
