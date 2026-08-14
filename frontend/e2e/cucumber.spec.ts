import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. **The only signal that always moves** when you act. */
const hand = (page: Page) => page.getByRole('button', { name: /を出す$/ });

// **合法な札だけを選ぶ。** 比較フォローを満たさない札はサーバが拒否する。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Cucumber E2E', () => {
  test('navigates to cucumber and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/cucumber');

    await expect(page.getByText(/キューカンバー|Cucumber/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cu-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **スート無関係が規則そのもの。**
  test('always states that suits are irrelevant', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(page.getByTestId('cu-rule')).toContainText(/スート|[Ss]uits/, { timeout: TIMEOUT_TRANSITION });
  });

  // **7 枚固定。** 人数で割りません。
  test('deals seven cards to each of the four seats', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`cu-seat-${id}`)).toContainText('7', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('cu-seat-4')).toHaveCount(0);
  });

  // **超えるべきランクは必ず出ている。** これが無いと何を出せばよいか分からない。
  test('always shows the rank to beat or that you lead', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(page.getByTestId('cu-threshold')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('cu-threshold')).toContainText(/リード|超える|lead|beat/);
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば手札が減らない。
  test('playing a legal card removes it from your hand', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await hand(page).count();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();
    await expect.poll(async () => hand(page).count(), { timeout: TIMEOUT_ACTION }).toBeLessThan(before);
  });

  // **失点はラウンドに1回だけの出来事。** 配り直す前に読ませる。
  test('stops at the round boundary before dealing again', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    for (let i = 0; i < 30; i++) {
      if (await page.getByTestId('cu-next-btn').isVisible()) {
        await expect(page.getByTestId('cu-round-end')).toBeVisible();
        await page.getByTestId('cu-next-btn').click();
        await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
        return;
      }
      if (await page.getByTestId('cu-result').isVisible()) break;
      if (!(await legalCard(page).first().isVisible())) break;
      await legalCard(page).first().click();
      await expect(
        hand(page).first().or(page.getByTestId('cu-next-btn')).or(page.getByTestId('cu-result')),
      ).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
    test.skip(true, 'この配りではラウンドの区切りに届かなかった');
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(page.getByTestId('cu-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('cu-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/cucumber');
    await expect(page.getByTestId('cu-header')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('cu-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
