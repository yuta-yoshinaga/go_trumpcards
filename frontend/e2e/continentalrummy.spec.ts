import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The human's hand cards.
 *
 * **`data-legal` is what marks a playable card**, and the hand section is found
 * by `data-tutorial` — a locator on the wrong attribute matches nothing and the
 * test then "passes" having clicked nothing.
 */
function handCards(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('[data-tutorial="continentalrummy-player-hand"] button[data-legal]');
}

/** The two draw buttons, keyed by shortcut rather than by label. */
function stockButton(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('button[aria-keyshortcuts="s"]');
}
function takeButton(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('button[aria-keyshortcuts="t"]');
}

test.describe('Continental Rummy E2E', () => {
  test('deals fifteen each and opens on the human draw', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    await expect(page.getByTestId('cont-layouts')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`cont-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
    await expect(page.getByTestId('cont-seat-1')).toContainText('15');
    await expect(page.getByTestId('cont-stock')).toBeVisible();
  });

  // **上がれる形は常に見えていること。** 5+5+5 がそこに無いのが肝。
  test('shows the three legal layouts and never advertises 5+5+5', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    const layouts = page.getByTestId('cont-layouts');
    await expect(layouts).toContainText('3+3+3+3+3', { timeout: TIMEOUT_TRANSITION });
    await expect(layouts).toContainText('4+4+4+3');
    await expect(layouts).toContainText('5+4+3+3');
    await expect(layouts).not.toContainText('5+5+5');
  });

  test('offers exactly one button for each draw', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    await expect(page.getByTestId('cont-layouts')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(stockButton(page)).toHaveCount(1);
    await expect(takeButton(page)).toHaveCount(1);
    // 引く前は捨てられない。
    await expect(handCards(page)).toHaveCount(0);
  });

  test('drawing from the stock opens the discard step', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    await expect(stockButton(page)).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const before = await page.getByTestId('cont-stock').innerText();
    await stockButton(page).click();
    await waitForLoaded(page);

    // 引いたら山が減り、捨てる札が押せるようになる。
    await expect(page.getByTestId('cont-stock')).not.toHaveText(before, { timeout: TIMEOUT_TRANSITION });
    await expect(handCards(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(stockButton(page)).toHaveCount(0);
  });

  test('discarding passes the turn and comes back to the draw', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    await expect(stockButton(page)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await stockButton(page).click();
    await waitForLoaded(page);

    const hand = handCards(page);
    await expect(hand.first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    await hand.first().click();
    await waitForLoaded(page);

    // 一周して自分の引く番に戻る (あるいはラウンドが決着している)。
    await expect(stockButton(page).or(page.getByTestId('cont-result')).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/continentalrummy');
    await expect(page.getByTestId('cont-layouts')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('cont-layouts')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
