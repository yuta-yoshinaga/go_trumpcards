import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Advances one round, whichever action the human owes.
 *
 * **The banker and a child owe different actions.** The page renders only the
 * one that applies, so click whichever is there.
 */
async function playRound(page: Parameters<typeof navigateTo>[0]) {
  const bet = page.getByTestId('kingo-bet');
  const deal = page.getByTestId('kingo-deal');
  if (await isVisibleWithin(deal, TIMEOUT_ACTION)) {
    await deal.click();
    await waitForLoaded(page);
    return;
  }
  if (await isVisibleWithin(bet, TIMEOUT_ACTION)) {
    await bet.click();
    await waitForLoaded(page);
  }
}

test.describe('Kingo E2E', () => {
  // **配る前は誰の手札も無い。** 隠しているのではなく、まだ存在しない。
  test('shows no cards until the bets are in', async ({ page }) => {
    await navigateTo(page, '/kingo');
    await expect(page.getByTestId('kingo-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // どの席にも手札が無い。
    await expect(page.getByTestId('kingo-cards-0')).toContainText('配る前です');
    await expect(page.getByTestId('kingo-rank-0')).toHaveCount(0);

    await playRound(page);

    // 決着すると全員ぶんが 3 枚ずつ出る。
    await expect(page.getByTestId('kingo-cards-0').locator('> *')).toHaveCount(3, {
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('kingo-rank-0')).toBeVisible();
  });

  // **親と子で求められる操作が違う。** 同時には出ない。
  test('asks for a bet or a deal, never both', async ({ page }) => {
    await navigateTo(page, '/kingo');
    await expect(page.getByTestId('kingo-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    for (let round = 0; round < 6; round++) {
      const betVisible = await isVisibleWithin(page.getByTestId('kingo-bet'), TIMEOUT_ACTION);
      const dealVisible = await isVisibleWithin(page.getByTestId('kingo-deal'), TIMEOUT_ACTION);
      if (!betVisible && !dealVisible) break;
      // どちらか一方だけ。
      expect(betVisible).not.toBe(dealVisible);

      await playRound(page);
      const next = page.getByTestId('kingo-next');
      if (!(await isVisibleWithin(next, TIMEOUT_ACTION))) break;
      await next.click();
      await waitForLoaded(page);
    }
  });

  test('plays rounds and advances the round counter', async ({ page }) => {
    await navigateTo(page, '/kingo');
    await expect(page.getByTestId('kingo-round')).toContainText('1', { timeout: TIMEOUT_TRANSITION });

    await playRound(page);
    const next = page.getByTestId('kingo-next');
    await expect(next).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('kingo-round')).toContainText('2', { timeout: TIMEOUT_TRANSITION });
    // 次のラウンドでは手札が消えている (また配る前)。
    await expect(page.getByTestId('kingo-cards-0')).toContainText('配る前です');
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/kingo');
    await expect(page.getByTestId('kingo-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByTestId('kingo-round')).toContainText('1', { timeout: TIMEOUT_TRANSITION });
  });
});
