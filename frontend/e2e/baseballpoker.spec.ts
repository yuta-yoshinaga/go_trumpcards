import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Plays the human seat until the hand settles.
 *
 * **Three different asks can be waiting.** Check and call swap places with the
 * board, and a face-up 3 interrupts both with a buy-or-fold choice — the page
 * only ever renders the one that applies, so click whichever is there.
 */
async function playOutHand(page: Parameters<typeof navigateTo>[0]) {
  const check = page.getByTestId('bb-check');
  const call = page.getByTestId('bb-call');
  const pay = page.getByTestId('bb-pay');
  for (let i = 0; i < 40; i++) {
    if (await isVisibleWithin(pay, TIMEOUT_ACTION)) {
      await pay.click();
      await waitForLoaded(page);
      continue;
    }
    if (await isVisibleWithin(check, TIMEOUT_ACTION)) {
      await check.click();
      await waitForLoaded(page);
      continue;
    }
    if (await isVisibleWithin(call, TIMEOUT_ACTION)) {
      await call.click();
      await waitForLoaded(page);
      continue;
    }
    return;
  }
}

test.describe('Baseball Poker E2E', () => {
  test('deals stud-style and shows every seat its up-cards', async ({ page }) => {
    await navigateTo(page, '/baseballpoker');

    // 3rd ストリートは 3 枚 (ボーナスが出た席は 4 枚)。
    await expect(page.getByTestId('bb-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const handCount = await page.getByTestId('bb-hand').locator('> *').count();
    expect(handCount).toBeGreaterThanOrEqual(3);

    // **CPU の表札は見えている。** 伏せ札だけが裏向き。
    await expect(page.getByTestId('bb-seat-cards-1')).toBeVisible();
    await expect(page.getByTestId('bb-seat-cards-1-hidden-0')).toBeVisible();
    await expect(page.getByTestId('bb-seat-cards-1-hidden-1')).toBeVisible();
    // 表札の位置は裏向きになっていない。
    await expect(page.getByTestId('bb-seat-cards-1-hidden-2')).toHaveCount(0);

    await expect(page.getByTestId('bb-street')).toContainText('1');
  });

  // **ワイルドとイベントの説明が画面に出ている。** 知らないと降りどころが決まらない。
  test('explains the wilds and the two events', async ({ page }) => {
    await navigateTo(page, '/baseballpoker');
    await expect(page.getByTestId('bb-wild-notice')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('bb-wild-notice')).toContainText('3');
    await expect(page.getByTestId('bb-wild-notice')).toContainText('9');
  });

  test('plays a hand out and starts the next one', async ({ page }) => {
    await navigateTo(page, '/baseballpoker');
    await expect(page.getByTestId('bb-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await playOutHand(page);

    const next = page.getByRole('button', { name: '次のハンドへ' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await next.click();
    await waitForLoaded(page);

    // 次のハンドではストリートが 1 に戻る。
    await expect(page.getByTestId('bb-street')).toContainText('1', { timeout: TIMEOUT_TRANSITION });
  });

  // **買い増しの返事はベットの手と同時に出ない。** 打ち間違いで払わないため。
  test('never shows the buy-in answer next to the betting actions', async ({ page }) => {
    await navigateTo(page, '/baseballpoker');
    await expect(page.getByTestId('bb-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const pay = page.getByTestId('bb-pay');
    for (let i = 0; i < 40; i++) {
      if (await isVisibleWithin(pay, TIMEOUT_ACTION)) {
        // 買い増しの場面では、ベットの手が出ていない。
        await expect(page.getByTestId('bb-check')).toHaveCount(0);
        await expect(page.getByTestId('bb-call')).toHaveCount(0);
        await expect(page.getByTestId('bb-buyfold')).toBeVisible();
        await pay.click();
        await waitForLoaded(page);
        continue;
      }
      const check = page.getByTestId('bb-check');
      const call = page.getByTestId('bb-call');
      if (await isVisibleWithin(check, TIMEOUT_ACTION)) {
        // ベット中は買い増しのボタンが出ていない。
        await expect(pay).toHaveCount(0);
        await check.click();
        await waitForLoaded(page);
        continue;
      }
      if (await isVisibleWithin(call, TIMEOUT_ACTION)) {
        await expect(pay).toHaveCount(0);
        await call.click();
        await waitForLoaded(page);
        continue;
      }
      break;
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/baseballpoker');
    await expect(page.getByTestId('bb-hand-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByTestId('bb-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
