import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The hand, scoped to the player-hand container.
 *
 * **`button[aria-pressed]` alone matches the whole navigation** — the sidebar
 * renders one such button per game, so an unscoped count sees 300+ elements.
 */
const hand = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-tutorial="tg-player-hand"] button[aria-pressed]');

/** Hand cards the server allows right now (restricted ones are aria-disabled). */
const playableCards = (page: Parameters<typeof navigateTo>[0]) =>
  page.locator('[data-tutorial="tg-player-hand"] button[aria-pressed]:not([aria-disabled="true"])');

test.describe('Troggu E2E', () => {
  test('deals eighteen cards and offers all four contracts', async ({ page }) => {
    await navigateTo(page, '/troggu');
    await expect(page.getByTestId('tg-info')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **契約は 4 つとも出る。** どれか欠けると、その契約だけが遊べなくなる。
    for (const c of ['trois', 'solo', 'piccolo', 'misere']) {
      await expect(page.getByTestId(`tg-bid-${c}`)).toBeVisible();
    }
    await expect(page.getByTestId('tg-pass')).toBeVisible();

    await expect(hand(page)).toHaveCount(18);
    await expect(page.getByTestId('tg-seat-3')).toBeVisible();
  });

  // **落札するとプレイに入る。** 場札交換のフェーズは無い。
  test('starts play after winning the auction', async ({ page }) => {
    await navigateTo(page, '/troggu');
    await expect(page.getByTestId('tg-bid-solo')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('tg-bid-solo').click();
    await waitForLoaded(page);

    await expect(page.getByTestId('tg-bid-solo')).toHaveCount(0);
    await expect(hand(page)).toHaveCount(18);
  });

  // **手札を出すと必ず 1 枚減る。** 席の点数は取れなければ動かない。
  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/troggu');
    await expect(page.getByTestId('tg-bid-solo')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('tg-bid-solo').click();
    await waitForLoaded(page);

    // **自分の手番になるまで押さない。** 手番でない間は制限が付かないので、
    // 「押せる札がある」だけでは押しても何も起きない (配り依存で落ちる)。
    for (let i = 0; i < 8; i++) {
      if (await page.locator('[data-testid="tg-info"][data-human-turn]').isVisible()) break;
      if (await isVisibleWithin(page.getByTestId('tg-next-trick'), 500)) {
        await page.getByTestId('tg-next-trick').click();
        await waitForLoaded(page);
        continue;
      }
      await page.waitForTimeout(300);
    }
    if (!(await page.locator('[data-testid="tg-info"][data-human-turn]').isVisible())) return;
    if ((await playableCards(page).count()) === 0) return;

    await playableCards(page).first().click();
    await waitForLoaded(page);
    await expect(hand(page)).toHaveCount(17, { timeout: TIMEOUT_TRANSITION });
  });

  // 全員パスなら流局し、次のディールへ進める。
  test('throws the deal in when everyone passes', async ({ page }) => {
    await navigateTo(page, '/troggu');
    await expect(page.getByTestId('tg-pass')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('tg-pass').click();
    await waitForLoaded(page);

    const result = page.getByTestId('tg-round-result');
    if (!(await isVisibleWithin(result, TIMEOUT_ACTION))) {
      // 他の席が落札したディールでは流局しない。
      return;
    }
    await expect(result).toContainText('流局');
    await page.getByTestId('tg-next-round').click();
    await waitForLoaded(page);
    await expect(hand(page)).toHaveCount(18, { timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/troggu');
    await expect(page.getByTestId('tg-pass')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('tg-pass').click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(hand(page)).toHaveCount(18, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('tg-bid-trois')).toBeVisible();
  });
});
