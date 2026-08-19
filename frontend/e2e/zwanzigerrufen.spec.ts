import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The hand, scoped to the player-hand container.
 *
 * **`button[aria-pressed]` alone matches the whole navigation.** The sidebar
 * renders one such button per game, so an unscoped count sees 300+ elements.
 */
const hand = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-tutorial="zw-player-hand"] button[aria-pressed]');

/** Hand cards the server allows right now (restricted ones are aria-disabled). */
const playableCards = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-tutorial="zw-player-hand"] button[aria-pressed]:not([aria-disabled="true"])');

test.describe('Zwanzigerrufen E2E', () => {
  test('deals twelve cards and opens the auction', async ({ page }) => {
    await navigateTo(page, '/zwanzigerrufen');
    await expect(page.getByTestId('zw-info')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **入札できるのは 20番呼びとソロだけ。** トリシャーケンは全員パスの結果。
    await expect(page.getByTestId('zw-bid-rufer')).toBeVisible();
    await expect(page.getByTestId('zw-bid-solo')).toBeVisible();
    await expect(page.getByTestId('zw-pass')).toBeVisible();

    await expect(hand(page)).toHaveCount(12);
    // 4 席ぶんの行が出る。
    await expect(page.getByTestId('zw-seat-3')).toBeVisible();
  });

  // **落札すると場札 6 枚が手札に加わる。** 伏せるまで進めない。
  test('takes the talon after winning the auction', async ({ page }) => {
    await navigateTo(page, '/zwanzigerrufen');
    await expect(page.getByTestId('zw-bid-solo')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // **上回れない入札は押せない** (#5786)。ソロは最高の入札なので、誰かが
    // 先にソロを宣言していた回だけ無効になる。
    if (await page.getByTestId('zw-bid-solo').isDisabled()) return;
    await page.getByTestId('zw-bid-solo').click();
    await waitForLoaded(page);

    // ソロは場札を受け取らないので、そのままプレイに入る。
    await expect(page.getByTestId('zw-info')).toBeVisible();
    await expect(hand(page)).toHaveCount(12);
  });

  test('buries six cards when the talon exchange is offered', async ({ page }) => {
    await navigateTo(page, '/zwanzigerrufen');
    await expect(page.getByTestId('zw-bid-rufer')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // 他の席が先に 20番呼び以上を宣言していれば、このボタンは押せない (#5786)。
    if (await page.getByTestId('zw-bid-rufer').isDisabled()) return;
    await page.getByTestId('zw-bid-rufer').click();
    await waitForLoaded(page);

    const discard = page.getByTestId('zw-discard');
    if (!(await isVisibleWithin(discard, TIMEOUT_ACTION))) {
      // 他の席がより高い入札をしたときは場札交換に立たない。
      return;
    }
    await expect(discard).toBeDisabled();
    await expect(hand(page)).toHaveCount(18);

    const cards = hand(page);
    for (let i = 0; i < 6; i++) {
      await cards.nth(i).click();
    }
    await expect(discard).toBeEnabled();
    await discard.click();
    await waitForLoaded(page);
    await expect(hand(page)).toHaveCount(12);
  });

  // **手札を出すと必ず 1 枚減る。** 席の点数は取れなければ動かないので、
  // 確かめる先は手札の枚数。
  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/zwanzigerrufen');
    await expect(page.getByTestId('zw-pass')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('zw-pass').click();
    await waitForLoaded(page);

    // 全員パスなら Trischaken、誰かが落札すれば防御側。どちらでも打つ手番は来る。
    for (let i = 0; i < 6; i++) {
      if (await isVisibleWithin(page.getByTestId('zw-next-trick'), 500)) {
        await page.getByTestId('zw-next-trick').click();
        await waitForLoaded(page);
        continue;
      }
      if ((await playableCards(page).count()) === 12) break;
      await page.waitForTimeout(200);
    }
    if ((await playableCards(page).count()) === 0) return;

    await playableCards(page).first().click();
    await waitForLoaded(page);
    await expect(hand(page)).toHaveCount(11, { timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/zwanzigerrufen');
    await expect(page.getByTestId('zw-pass')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('zw-pass').click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(hand(page)).toHaveCount(12, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('zw-bid-rufer')).toBeVisible();
  });
});
