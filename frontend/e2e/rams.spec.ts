import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Rams E2E', () => {
  test('navigates to rams and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/rams');

    await expect(page.getByText(/ラムス|Rams/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **ポット・切り札・リスクの3点が参加判断の材料。** 盤面からは読めない。
  test('shows the pot, the trump card and the penalty', async ({ page }) => {
    await navigateTo(page, '/rams');
    // 開始時のポットは 4 人 × アンティ 3。
    await expect(page.getByTestId('rm-pot')).toContainText('12', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rm-trump')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rm-risk')).toContainText('5', { timeout: TIMEOUT_TRANSITION });
  });

  // **配り直後は参加選択。** 両方のボタンが実サーバ相手に出ていること。
  test('opens on the decision phase with both choices', async ({ page }) => {
    await navigateTo(page, '/rams');
    await expect(page.getByTestId('rm-in-btn')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('rm-out-btn')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('entering the round starts play and hides the choices', async ({ page }) => {
    await navigateTo(page, '/rams');
    await page.getByTestId('rm-in-btn').click();

    await expect(page.getByTestId('rm-in-btn')).toHaveCount(0, { timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('rm-seat-0')).toContainText(/参加|in/, { timeout: TIMEOUT_ACTION });
  });

  // **降りたラウンドは「見ている」と伝える。** 操作待ちに見えてはいけない。
  test('dropping out says it is watching', async ({ page }) => {
    await navigateTo(page, '/rams');
    await page.getByTestId('rm-out-btn').click();

    await expect(page.getByTestId('rm-seat-0')).toContainText(/降り|out/, { timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the trick counter advances', async ({ page }) => {
    await navigateTo(page, '/rams');
    await page.getByTestId('rm-in-btn').click();
    await expect(page.getByTestId('rm-trick')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('rm-trick').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(async () => page.getByTestId('rm-trick').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/rams');
    await expect(page.getByTestId('rm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('rm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/rams');
    await expect(page.getByTestId('rm-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByRole('button', { name: /^投了$|^Give up$/ })).toHaveCount(0, { timeout: TIMEOUT_ACTION });
  });
});
