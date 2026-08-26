import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Selects the first selectable card in the human's hand.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — it carries no `data-testid`. A locator on the
 * wrong attribute matches nothing, and the test then "passes" having clicked
 * nothing at all.
 */
async function selectFirstCard(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  const cards = page.locator('[data-tutorial="dehlapakad-player-hand"] button[data-legal]');
  if ((await cards.count()) === 0) return false;
  await cards.first().click();
  return true;
}

test.describe('Dehla Pakad E2E', () => {
  // **開幕は人間が切り札を宣言する。** 席 3 が親なので、その右隣の席 0 が決める。
  test('opens with the human calling the trump from five cards', async ({ page }) => {
    await navigateTo(page, '/dehlapakad');
    await expect(page.getByText('ハンド 1（2 コート先取）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const choices = page.getByTestId('dehlapakad-trump-choices');
    await expect(choices).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(choices.locator('button')).toHaveCount(4);
    // 宣言前は切り札もトリック番号も出ない。
    await expect(page.getByTestId('dehlapakad-trump')).toHaveCount(0);
  });

  test('calls a trump and plays a card', async ({ page }) => {
    await navigateTo(page, '/dehlapakad');
    await expect(page.getByText('ハンド 1（2 コート先取）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByTestId('dehlapakad-trump-3').click();
    await waitForLoaded(page);
    await expect(page.getByTestId('dehlapakad-trump')).toContainText('ハート', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('dehlapakad-trump-choices')).toHaveCount(0);

    const play = page.getByTestId('dehlapakad-play');
    if (!(await isVisibleWithin(play, TIMEOUT_ACTION))) return; // まだ CPU の手番
    await expect(play).toBeDisabled();
    expect(await selectFirstCard(page)).toBe(true);
    await expect(play).toBeEnabled();
    await play.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('dehlapakad-next-hand').or(play).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  // **取っても札はもらえない。** 1 トリック目が片付いた時点で、中央に山が
  // 残っているのが正しい姿。
  test('leaves the first trick in the centre rather than handing it over', async ({ page }) => {
    await navigateTo(page, '/dehlapakad');
    await expect(page.getByText('ハンド 1（2 コート先取）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('dehlapakad-trump-3').click();
    await waitForLoaded(page);

    const play = page.getByTestId('dehlapakad-play');
    if (!(await isVisibleWithin(play, TIMEOUT_ACTION))) return;
    expect(await selectFirstCard(page)).toBe(true);
    await play.click();
    await waitForLoaded(page);

    // 1 トリック目を片付けた直後、山は必ず残っている (2 連勝は成立しえない)。
    const pile = page.getByTestId('dehlapakad-centre-pile');
    await expect(pile).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(pile).toContainText('中央に');
    await expect(page.getByTestId('dehlapakad-pile-goes-to')).toBeVisible();
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/dehlapakad');
    await expect(page.getByText('ハンド 1（2 コート先取）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ハンド 1（2 コート先取）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
