import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The human's selectable hand cards.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — a locator on the wrong attribute matches nothing
 * and the test then "passes" having clicked nothing.
 */
function handCards(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('[data-tutorial="diloti-player-hand"] button[data-legal]');
}

test.describe('Diloti E2E', () => {
  test('deals six each with four on the table', async ({ page }) => {
    await navigateTo(page, '/diloti');
    await expect(page.getByText('局 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **場札には番号が要る。** 取る対象はこの番号で指す。
    const table = page.getByTestId('diloti-table');
    await expect(table).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(table.getByRole('img')).toHaveCount(4);

    await expect(handCards(page)).toHaveCount(6);
    // 山は 52 - 6*2 - 4 = 36 枚。
    await expect(page.getByTestId('diloti-deck')).toContainText('36');
  });

  // **打てる手はサーバが数えたものだけ。** 札を選ぶと候補が出る。
  test('shows the move options for the selected card', async ({ page }) => {
    await navigateTo(page, '/diloti');
    await expect(page.getByText('局 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 選ぶ前は候補が出ない。
    await expect(page.getByTestId('diloti-move-options')).toHaveCount(0);
    await handCards(page).first().click();
    await expect(page.getByTestId('diloti-move-options')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('plays a card and advances the board', async ({ page }) => {
    await navigateTo(page, '/diloti');
    await expect(page.getByText('局 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const hand = handCards(page);
    await expect(hand).toHaveCount(6);
    await hand.first().click();

    // **選べば必ず打つ手がある。** 取れる組も宣言も無ければ場に置く手が出るので、
    // ここは「候補が出たら」ではなく無条件に 1 手打てる ── 条件で包むと、
    // 一度も打たないまま緑になる。
    const box = page.getByTestId('diloti-move-options');
    await expect(box).toBeVisible({ timeout: TIMEOUT_ACTION });
    const buttons = box.locator('button');
    await expect(buttons.first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await buttons.first().click();
    await waitForLoaded(page);

    // 打てば手札が 1 枚減る (局が切れたら区切りの案内が出る)。
    await expect(
      hand.or(page.getByTestId('diloti-next-round')).or(page.getByTestId('diloti-winner')).first(),
    ).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    if ((await page.getByTestId('diloti-next-round').count()) === 0) {
      await expect(hand).toHaveCount(5);
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/diloti');
    await expect(page.getByText('局 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('局 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
