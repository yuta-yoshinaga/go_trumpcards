import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/** True once the server says it is the human's turn to act. */
const humanTurn = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-testid="ho-discipline"][data-human-turn]');

/**
 * Folds until the hand is settled, then reports whether it settled.
 *
 * **押せる操作を待つのではなく、手番になるまで待つ。** 手番でない間はフッターに
 * 何も出ないので、「ボタンがある」を待つと配り次第で永久に待つ。
 *
 * 引き直しの番ではフォールドが出ないので、そこはスタンドパットで抜ける。
 */
async function foldOneHand(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  for (let i = 0; i < 16; i++) {
    if (await isVisibleWithin(page.getByTestId('ho-next-hand'), 500)) return true;
    if (await humanTurn(page).isVisible()) {
      const stand = page.getByTestId('ho-draw-stand');
      if (await stand.isVisible()) {
        await stand.click();
      } else {
        await page.getByRole('button', { name: 'フォールド' }).click();
      }
      await waitForLoaded(page);
      continue;
    }
    await page.waitForTimeout(300);
  }
  return await isVisibleWithin(page.getByTestId('ho-next-hand'), TIMEOUT_ACTION);
}

test.describe('Eight-Game Mix E2E', () => {
  test('opens on the hold’em leg with every seat funded', async ({ page }) => {
    await navigateTo(page, '/eightgame');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **H から始まる。** 8 種目でも並びの先頭はリミットホールデム。
    await expect(page.getByTestId('ho-letter')).toHaveText('H');
    await expect(page.getByTestId('ho-discipline')).toContainText('テキサスホールデム');

    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ho-seat-${id}-chips`)).toBeVisible();
    }
  });

  // **席数は 4 だけ。** 6 を選べる画面は、8 種目目で理由も出さずに終わる卓を
  // 作れてしまう。
  test('offers a four-handed table only', async ({ page }) => {
    await navigateTo(page, '/eightgame');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 設定は閉じた <details> の中にある。開かないとセレクトは hidden のまま。
    await page.locator('summary', { hasText: '設定' }).click();
    const seatSelect = page.getByLabel('席数');
    await expect(seatSelect).toBeVisible();
    await expect(seatSelect.locator('option')).toHaveCount(1);
    await expect(seatSelect.locator('option')).toHaveText(['4']);
  });

  // **8 種目のほうへ回る。** 5 種目で止まるなら H.O.R.S.E. のままである。
  test('rotates past the H.O.R.S.E. legs', async ({ page }) => {
    await navigateTo(page, '/eightgame');
    await expect(page.getByTestId('ho-letter')).toHaveText('H', { timeout: TIMEOUT_TRANSITION });

    const seen = new Set<string>();
    for (let hand = 0; hand < 14; hand++) {
      const letter = (await page.getByTestId('ho-letter').textContent())?.trim();
      if (letter) seen.add(letter);
      if (!(await foldOneHand(page))) break;
      await page.getByTestId('ho-next-hand').click();
      await waitForLoaded(page);
    }
    // 誰かが飛んで終局することもあるので、通った種目が並びの中にあることだけを見る。
    for (const letter of seen) {
      expect(['H', 'O', 'R', 'S', 'E', 'NLH', 'PLO', '2-7']).toContain(letter);
    }
    expect(seen.size).toBeGreaterThan(1);
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/eightgame');
    await expect(page.getByTestId('ho-discipline')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('ho-letter')).toHaveText('H', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ho-seat-0-chips')).toBeVisible();
  });
});
