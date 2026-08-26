import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Selects the first selectable card in the human's hand.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — a locator on the wrong attribute matches nothing
 * and the test then "passes" having clicked nothing.
 */
async function selectFirstCard(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  const cards = page.locator('[data-tutorial="cirulla-player-hand"] button[data-legal]');
  if ((await cards.count()) === 0) return false;
  await cards.first().click();
  return true;
}

test.describe('Cirulla E2E', () => {
  test('deals three each with four on the table', async ({ page }) => {
    await navigateTo(page, '/cirulla');
    await expect(page.getByText('ラウンド 1（51 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **場札には番号が要る。** 取る札はこの番号で指す。
    const table = page.getByTestId('cirulla-table');
    await expect(table).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(table.getByRole('img')).toHaveCount(4);

    // 山は 40 - 3*2 - 4 = 30 枚。
    await expect(page.getByTestId('cirulla-deck')).toContainText('30');
  });

  // **取れる組はサーバが数えたものだけ。** 札を選ぶと候補が出る。
  test('shows the capture choices for the selected card', async ({ page }) => {
    await navigateTo(page, '/cirulla');
    await expect(page.getByText('ラウンド 1（51 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 選ぶ前は候補が出ない。
    await expect(page.getByTestId('cirulla-capture-options')).toHaveCount(0);
    expect(await selectFirstCard(page)).toBe(true);
    await expect(page.getByTestId('cirulla-capture-options')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('plays a card and advances the board', async ({ page }) => {
    await navigateTo(page, '/cirulla');
    await expect(page.getByText('ラウンド 1（51 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const hand = page.locator('[data-tutorial="cirulla-player-hand"] button[data-legal]');
    await expect(hand).toHaveCount(3);
    await hand.first().click();

    // **選べば必ず打つ手がある。** 取れる組が無ければ場に置く手が出るので、
    // ここは「候補が出たら」ではなく無条件に 1 手打てる ── 条件で包むと、
    // 一度も打たないまま緑になる。
    const box = page.getByTestId('cirulla-capture-options');
    await expect(box).toBeVisible({ timeout: TIMEOUT_ACTION });
    const buttons = box.locator('button');
    await expect(buttons.first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await buttons.first().click();
    await waitForLoaded(page);

    // 打てば手札が 1 枚減る (ラウンドが切れたら区切りの案内が出る)。
    await expect(
      hand.or(page.getByTestId('cirulla-next-round')).or(page.getByTestId('cirulla-winner')).first(),
    ).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    if ((await page.getByTestId('cirulla-next-round').count()) === 0) {
      await expect(hand).toHaveCount(2);
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/cirulla');
    await expect(page.getByText('ラウンド 1（51 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ラウンド 1（51 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
