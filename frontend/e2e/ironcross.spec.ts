import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Plays the human seat until the table asks which arm of the cross to use.
 *
 * **Check and call swap places depending on the board**, so neither is always
 * legal — the page only renders the one that applies. Click whichever is there.
 */
async function betUntilChoose(page: Parameters<typeof navigateTo>[0]) {
  const check = page.getByTestId('ic-check');
  const call = page.getByTestId('ic-call');
  for (let i = 0; i < 30; i++) {
    if (await isVisibleWithin(page.getByTestId('ic-vertical'), TIMEOUT_ACTION)) return;
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

/** Plays the hand out, taking the vertical arm when the choice comes up. */
async function playOutHand(page: Parameters<typeof navigateTo>[0]) {
  await betUntilChoose(page);
  const vertical = page.getByTestId('ic-vertical');
  if (await isVisibleWithin(vertical, TIMEOUT_ACTION)) {
    await vertical.click();
    await waitForLoaded(page);
  }
}

test.describe('Iron Cross E2E', () => {
  test('deals four cards and opens the cross one at a time', async ({ page }) => {
    await navigateTo(page, '/ironcross');

    // **手札は 4 枚。** ホールデムの 2 枚でもシンシナティの 5 枚でもない。
    await expect(page.getByTestId('ic-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ic-hand').locator('> *')).toHaveCount(4);
    await expect(page.getByTestId('ic-revealed')).toContainText('0');

    // **十字の枠は最初から 5 つある。** 伏せている位置も落とさない。
    for (const i of [0, 1, 2, 3, 4]) {
      await expect(page.getByTestId(`ic-cross-${i}`)).toBeVisible();
    }

    await betUntilChoose(page);
    await expect(page.getByTestId('ic-revealed')).toContainText('5', { timeout: TIMEOUT_TRANSITION });
  });

  // **縦横の選択はベットの手とは別に出る。** 一度きりで取り返しがつかない。
  test('asks which arm to play once the cross is complete', async ({ page }) => {
    await navigateTo(page, '/ironcross');
    await expect(page.getByTestId('ic-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 配った直後は選ぶボタンが無い。
    await expect(page.getByTestId('ic-vertical')).toHaveCount(0);

    await betUntilChoose(page);
    await expect(page.getByTestId('ic-vertical')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ic-horizontal')).toBeVisible();
    // 選ぶ場面ではベットの手を出さない。
    await expect(page.getByTestId('ic-check')).toHaveCount(0);

    await page.getByTestId('ic-vertical').click();
    await waitForLoaded(page);
    // 押した列がそのまま自分の席に残る。
    await expect(page.getByTestId('ic-line-0')).toContainText('縦', { timeout: TIMEOUT_TRANSITION });
  });

  // **CPU の手札はショーダウンまでサーバが送らない。**
  test('cpu hands stay hidden until showdown', async ({ page }) => {
    await navigateTo(page, '/ironcross');
    await expect(page.getByTestId('ic-seat-cards-1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ic-seat-cards-1')).toContainText('伏せ');

    await playOutHand(page);

    await expect(page.getByTestId('ic-seat-cards-1')).not.toContainText('伏せ', {
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('plays a hand out and starts the next one', async ({ page }) => {
    await navigateTo(page, '/ironcross');
    await expect(page.getByTestId('ic-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await playOutHand(page);

    const next = page.getByRole('button', { name: '次のハンドへ' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await next.click();
    await waitForLoaded(page);

    // 次のハンドでは十字が 0 枚に戻り、手札が配り直される。
    await expect(page.getByTestId('ic-revealed')).toContainText('0', { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('ic-hand').locator('> *')).toHaveCount(4);
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/ironcross');
    await expect(page.getByTestId('ic-hand-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByTestId('ic-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
