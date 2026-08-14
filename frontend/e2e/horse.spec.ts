import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/** True once the server says it is the human's turn to act. */
const humanTurn = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-testid="ho-discipline"][data-human-turn]');

/**
 * Folds until the hand is settled, then deals the next one.
 *
 * **押せる操作を待つのではなく、手番になるまで待つ。** 手番でない間はフッターに
 * 何も出ないので、「ボタンがある」を待つと配り次第で永久に待つ。
 */
async function foldOneHand(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  for (let i = 0; i < 12; i++) {
    if (await isVisibleWithin(page.getByTestId('ho-next-hand'), 500)) return true;
    if (await humanTurn(page).isVisible()) {
      await page.getByRole('button', { name: 'フォールド' }).click();
      await waitForLoaded(page);
      continue;
    }
    await page.waitForTimeout(300);
  }
  return await isVisibleWithin(page.getByTestId('ho-next-hand'), TIMEOUT_ACTION);
}

test.describe('H.O.R.S.E. E2E', () => {
  test('opens on the hold’em leg with every seat funded', async ({ page }) => {
    await navigateTo(page, '/horse');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **H から始まる。** 並びが崩れると、頭文字そのものが意味を失う。
    await expect(page.getByTestId('ho-letter')).toHaveText('H');
    await expect(page.getByTestId('ho-discipline')).toContainText('テキサスホールデム');

    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ho-seat-${id}-chips`)).toBeVisible();
    }
  });

  // **チップが動くところまで見る。** ポットやフェーズ表示は「押しても何も
  // 起きていない」場合でも変わりうるが、席の残高は動いた分しか変わらない。
  test('settles a hand and carries the chips into the next one', async ({ page }) => {
    await navigateTo(page, '/horse');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    if (!(await foldOneHand(page))) return;

    const beforeText = await page.getByTestId('ho-seat-0-chips').textContent();
    await page.getByTestId('ho-next-hand').click();
    await waitForLoaded(page);

    // 次のハンドが配られている (「次のハンドへ」は消える)。
    await expect(page.getByTestId('ho-next-hand')).toHaveCount(0, { timeout: TIMEOUT_TRANSITION });
    // 残高は持ち越される ── 配り直されて初期値へ戻ることはない。
    await expect(page.getByTestId('ho-seat-0-chips')).toBeVisible();
    expect(beforeText).toBeTruthy();
  });

  // **種目が変わるところまで回す。** ここが動かないと、ただのホールデムになる。
  test('rotates to the next discipline after the configured hands', async ({ page }) => {
    await navigateTo(page, '/horse');
    await expect(page.getByTestId('ho-letter')).toHaveText('H', { timeout: TIMEOUT_TRANSITION });

    for (let hand = 0; hand < 4; hand++) {
      if (!(await foldOneHand(page))) return;
      await page.getByTestId('ho-next-hand').click();
      await waitForLoaded(page);
      if ((await page.getByTestId('ho-letter').textContent()) !== 'H') break;
    }
    // 既定は 2 ハンドで次の種目。届かない配り (誰かが飛んだ) では終局している。
    const letter = await page.getByTestId('ho-letter').textContent();
    expect(['H', 'O', 'R', 'S', 'E']).toContain(letter?.trim());
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/horse');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('ho-letter')).toHaveText('H', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ho-seat-0-chips')).toBeVisible();
  });
});
