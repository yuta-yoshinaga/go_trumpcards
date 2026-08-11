import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. Cards are selected first, then a capture option is chosen. */
const handCard = (page: Page) => page.getByRole('button', { name: /を選ぶ$/ });

/** A card the server says can capture — those get the green ring. */
const capturingCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Pasur E2E', () => {
  test('navigates to pasur and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/pasur');

    await expect(page.getByText(/パスール|Pasur/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ps-pack')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **11 の合計と絵札の扱いが規則そのもの。**
  test('always states the capture rule', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(page.getByTestId('ps-rule')).toContainText(/11/, { timeout: TIMEOUT_TRANSITION });
  });

  test('shows every seat', async ({ page }) => {
    await navigateTo(page, '/pasur');
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ps-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('ps-seat-4')).toHaveCount(0);
  });

  // **選ぶと、サーバが送った取り方だけが出る。**
  test('offers capture options once a card is selected', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(handCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    await handCard(page).first().click();
    await expect(page.getByTestId('ps-options')).toBeVisible({ timeout: TIMEOUT_ACTION });
    // 取れる札なら取り方が、取れない札なら「場に置く」が出る——どちらかは必ず出る。
    const takeOrTrail = page.locator('[data-testid^="ps-take-"], [data-testid="ps-trail-btn"]');
    await expect(takeOrTrail.first()).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば手札が減らない。
  test('playing a card removes it from the hand', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(handCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await handCard(page).count();
    expect(before).toBeGreaterThan(0);

    // 取れる札があればそれを、無ければ先頭を選ぶ。
    const target = (await capturingCard(page).count()) > 0 ? capturingCard(page).first() : handCard(page).first();
    await target.click();
    await page.locator('[data-testid^="ps-take-"], [data-testid="ps-trail-btn"]').first().click();

    // **必ず動く信号は手札の枚数。** 取れたかどうかに依らない。
    await expect.poll(async () => handCard(page).count(), { timeout: TIMEOUT_ACTION }).toBeLessThan(before);
  });

  // **取れるときは場に置けない。** サーバが必ず拒否するので出さない。
  test('never offers to lay down a card that can capture', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(handCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    if ((await capturingCard(page).count()) === 0) {
      test.skip(true, 'この配りでは取れる札が無い');
    }
    await capturingCard(page).first().click();
    await expect(page.getByTestId('ps-must-capture')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('ps-trail-btn')).toHaveCount(0);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(page.getByTestId('ps-pack')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('ps-pack')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/pasur');
    await expect(page.getByTestId('ps-pack')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('ps-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
