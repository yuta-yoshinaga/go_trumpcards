import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Selects the first selectable card in the human's hand.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — it carries no `data-testid`. A locator on the
 * wrong attribute matches nothing and the test then "passes" having clicked
 * nothing at all.
 */
async function selectFirstCard(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  const cards = page.locator('[data-tutorial="quodlibet-player-hand"] button[data-legal]');
  if ((await cards.count()) === 0) return false;
  await cards.first().click();
  return true;
}

test.describe('Quodlibet E2E', () => {
  // **席 0 が第 1 ディールの親。** 開幕は必ず人間の種目選択から始まる。
  test('opens on the human dealer choosing from the first wheel', async ({ page }) => {
    await navigateTo(page, '/quodlibet');
    await expect(page.getByText('ディール 1/12')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText('第 1 の輪')).toBeVisible();

    const choices = page.getByTestId('quodlibet-contract-choices');
    await expect(choices).toBeVisible({ timeout: TIMEOUT_ACTION });
    // 選べるのはこの輪の 4 種目だけ。第 2 の輪 (id 4) は出ない。
    await expect(choices.locator('button')).toHaveCount(4);
    await expect(page.getByTestId('quodlibet-contract-4')).toHaveCount(0);

    // 点の向きを画面に出す ── 少ないほうが勝ち。
    await expect(page.getByText('罰点（少ないほうが勝ち）')).toBeVisible();
  });

  test('plays a deal of Minus through to a card being played', async ({ page }) => {
    await navigateTo(page, '/quodlibet');
    await expect(page.getByText('ディール 1/12')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // マイナス (id 1) を選ぶ。
    await page.getByTestId('quodlibet-contract-1').click();
    await waitForLoaded(page);
    await expect(page.getByTestId('quodlibet-contract')).toContainText('マイナス', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('quodlibet-contract-choices')).toHaveCount(0);

    const play = page.getByTestId('quodlibet-play');
    if (!(await isVisibleWithin(play, TIMEOUT_ACTION))) return; // まだ CPU の手番
    await expect(play).toBeDisabled();
    expect(await selectFirstCard(page)).toBe(true);
    await expect(play).toBeEnabled();
    await play.click();
    await waitForLoaded(page);

    // 盤面は進む: 次のディールか、次の自分の手番。
    await expect(page.getByTestId('quodlibet-next-deal').or(play).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  // **四分と小食いはトリックではない。** 場が出て、トリック番号は消える。
  test('shows a shed area rather than a trick for Snack', async ({ page }) => {
    await navigateTo(page, '/quodlibet');
    await expect(page.getByText('ディール 1/12')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 第 3 の輪に届くまで、各輪の残りを順に消化する。
    for (let deal = 0; deal < 12; deal++) {
      const choices = page.getByTestId('quodlibet-contract-choices');
      if (await choices.isVisible()) {
        await choices.locator('button').first().click();
        await waitForLoaded(page);
      }
      const shed = page.getByTestId('quodlibet-shed-area');
      if (await shed.isVisible()) {
        await expect(page.getByText(/トリック \d+\/8/)).toHaveCount(0);
        return;
      }
      const play = page.getByTestId('quodlibet-play');
      const next = page.getByTestId('quodlibet-next-deal');
      if (await next.isVisible()) {
        await next.click();
        await waitForLoaded(page);
        continue;
      }
      if (!(await play.isVisible())) break;
      const pass = page.getByTestId('quodlibet-pass');
      if (await pass.isVisible()) {
        await pass.click();
      } else {
        if (!(await selectFirstCard(page))) break;
        await play.click();
      }
      await waitForLoaded(page);
      deal--; // まだ同じディールの中
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/quodlibet');
    await expect(page.getByText('ディール 1/12')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ディール 1/12')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
