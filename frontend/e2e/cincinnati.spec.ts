import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Plays the human seat until the hand settles.
 *
 * **Check and call swap places depending on the board**, so neither is always
 * legal — the page only renders the one that applies. Click whichever is there.
 */
async function playOutHand(page: Parameters<typeof navigateTo>[0]) {
  const check = page.getByTestId('cin-check');
  const call = page.getByTestId('cin-call');
  for (let i = 0; i < 30; i++) {
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

test.describe('Cincinnati E2E', () => {
  test('deals five cards and reveals the board one at a time', async ({ page }) => {
    await navigateTo(page, '/cincinnati');

    // **手札は 5 枚。** ホールデムの 2 枚ではない。
    await expect(page.getByTestId('cin-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cin-hand').locator('> *')).toHaveCount(5);
    // 配った直後は場が 0 枚。
    await expect(page.getByTestId('cin-revealed')).toContainText('0');

    // 1 手進めると場が 1 枚増える (全員が動いたラウンドの後)。
    await playOutHand(page);
    await expect(page.getByTestId('cin-revealed')).toContainText('5', { timeout: TIMEOUT_TRANSITION });
  });

  // **CPU の手札はショーダウンまでサーバが送らない。**
  test('cpu hands stay hidden until showdown', async ({ page }) => {
    await navigateTo(page, '/cincinnati');
    await expect(page.getByTestId('cin-seat-cards-1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cin-seat-cards-1')).toContainText('伏せ');

    await playOutHand(page);

    // ショーダウンでは開く。
    await expect(page.getByTestId('cin-seat-cards-1')).not.toContainText('伏せ', {
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('plays a hand out and starts the next one', async ({ page }) => {
    await navigateTo(page, '/cincinnati');
    await expect(page.getByTestId('cin-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await playOutHand(page);

    const next = page.getByRole('button', { name: '次のハンドへ' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await next.click();
    await waitForLoaded(page);

    // 次のハンドでは場が 0 枚に戻り、手札が配り直される。
    await expect(page.getByTestId('cin-revealed')).toContainText('0', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cin-hand').locator('> *')).toHaveCount(5);
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/cincinnati');
    await expect(page.getByTestId('cin-hand-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByTestId('cin-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
