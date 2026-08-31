import { expect, test } from '@playwright/test';
import {
  isVisibleWithin,
  navigateTo,
  TIMEOUT_ACTION,
  TIMEOUT_GAME_LOOP,
  TIMEOUT_TRANSITION,
  waitForLoaded,
} from './helpers';

/**
 * The human's playable hand cards.
 *
 * **`data-legal` is what marks a playable card**, and the hand section is
 * found by `data-tutorial` — a locator on the wrong attribute matches nothing
 * and the test then "passes" having clicked nothing.
 */
function handCards(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('[data-tutorial="comet-player-hand"] button[data-legal]');
}

/**
 * Anything that means the human may act again.
 *
 * **`data-legal` disappears while the CPUs play** — the page only marks legal
 * cards on the human's turn (`canPlay = isPlayPhase && state.isHumanTurn`), and
 * a hand with no legal card shows the pass button instead. Waiting on the cards
 * alone therefore fails for any deal where the human cannot follow, which is
 * what made this spec fail 3 runs in 8 (#6836).
 */
function humanCanAct(page: Parameters<typeof navigateTo>[0]) {
  return handCards(page)
    .or(page.getByTestId('comet-pass'))
    .or(page.getByTestId('comet-next-round'))
    .or(page.getByTestId('comet-winner'))
    .first();
}

test.describe('Comet E2E', () => {
  test('deals the pack out and buries the remainder', async ({ page }) => {
    await navigateTo(page, '/comet');
    await expect(page.getByText('局 1（100 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 51 枚を 4 人に配ると 12 枚ずつ、余り 3 枚が死に手。
    await expect(page.getByTestId('comet-dead')).toContainText('3');
    await expect(handCards(page)).toHaveCount(12);

    // **先頭は何でも出せる。**
    await expect(page.getByTestId('comet-need')).toContainText('好きな札');
    // 連なりはまだ空。
    await expect(page.getByTestId('comet-pile')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the sequence names the next rank', async ({ page }) => {
    await navigateTo(page, '/comet');
    await expect(page.getByText('局 1（100 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const hand = handCards(page);
    await expect(hand).toHaveCount(12);
    // **先頭なので必ず打てる。** 条件で包むと、一度も打たないまま緑になる。
    await hand.first().click();
    await waitForLoaded(page);

    // **手番は 3 人の CPU を回って戻ってくる。**その間 `data-legal` は 1 つも
    // 付かないので、札だけを待つと出せない配りで必ず落ちる。パスも区切りも
    // 「人間が次に動ける」印なので同じ待ちに入れ、CPU が打ち切る時間を見る。
    await expect(humanCanAct(page), 'the human never got another turn after playing').toBeVisible({
      timeout: TIMEOUT_GAME_LOOP,
    });
    if ((await page.getByTestId('comet-next-round').count()) === 0) {
      await expect(page.getByTestId('comet-need')).toBeVisible({ timeout: TIMEOUT_ACTION });
      await expect(handCards(page).or(page.getByTestId('comet-pass')).first()).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
    }
  });

  test('shows every seat with its card count', async ({ page }) => {
    await navigateTo(page, '/comet');
    await expect(page.getByText('局 1（100 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    const scores = page.getByTestId('comet-scores');
    await expect(scores).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(scores).toContainText('手札 12 枚');
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/comet');
    await expect(page.getByText('局 1（100 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('局 1（100 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
