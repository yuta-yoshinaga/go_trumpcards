import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Put E2E', () => {
  test('starts a match: reset → verify game progresses → reset', async ({ page }) => {
    await navigateTo(page, '/put');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'プット' })).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/マッチ得点/)).toBeVisible({ timeout: 10_000 });

    // Reset again to confirm the match can be restarted from any phase.
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'プット' })).toBeVisible({ timeout: 10_000 });
  });

  test('the card-strength panel states Put’s order, not Truco’s', async ({ page }) => {
    await navigateTo(page, '/put');
    await waitForLoaded(page);

    // **This panel is the one place the divergence is visible to the player.**
    // Cloned from Truco it listed suit-specific matadores (1♠ > 1♣ > 7♠ > 7♦)
    // and said 8/9/10 are unused — both false for Put, and both would mislead
    // the reader about the only rule that makes this game unusual.
    const panel = page.getByTestId('put-rank-ref');
    await expect(panel).toBeVisible({ timeout: 10_000 });
    await panel.click(); // <details> — open it

    await expect(panel).toContainText('3 ＞ 2 ＞ A ＞ K');
    await expect(panel).not.toContainText('マタドール');
    await expect(panel).not.toContainText('8・9・10 は使いません');
  });

  test('plays a card and the hand keeps moving', async ({ page }) => {
    await navigateTo(page, '/put');
    await waitForLoaded(page);

    const hand = page.locator('[data-tutorial="put-hand"]');
    const cards = hand.getByRole('button');
    const nextButton = page.getByRole('button', { name: '次へ' });

    await expect(hand, 'the human hand must render').toBeVisible({ timeout: 10_000 });
    await expect(cards, 'a Put hand is exactly 3 cards').toHaveCount(3);

    // **Drive to the player's turn instead of assuming it.** Whoever is
    // non-dealer leads, so on some deals the CPU plays first and the hand sits
    // at trick end waiting for "次へ". On others the CPU declares Put and the
    // hand sits in the respond phase waiting for 受諾/拒否. The cards are
    // disabled in both, so assuming a playable turn on load is deal-dependent —
    // measured: 4 of 8 runs landed in one of these states.
    const acceptButton = page.getByRole('button', { name: '受諾' });
    for (let i = 0; i < 10; i++) {
      const playable = await cards
        .first()
        .isEnabled()
        .catch(() => false);
      if (playable) break;
      if (await nextButton.isVisible().catch(() => false)) {
        await nextButton.click();
      } else if (await acceptButton.isVisible().catch(() => false)) {
        // Accepting keeps the hand alive; declining would end it.
        await acceptButton.click();
      } else {
        break;
      }
      await waitForLoaded(page);
    }

    // **Assert we got there, unconditionally.** Wrapping the play in
    // `if (enabled)` would let this pass on every deal that never reached the
    // player's turn — the exact hole that let the Dramaha draw-round spec go
    // green without ever clicking its action button.
    await expect(cards.first(), 'never reached the human turn, so playing a card was never exercised').toBeEnabled({
      timeout: 10_000,
    });

    // Put has no must-follow, so the first card is always legal.
    await cards.first().click();
    await waitForLoaded(page);

    // The play landed: the hand is down a card, or the trick resolved and a
    // fresh hand was dealt. Either way the match header must still render —
    // the page must not be stuck on an error or a blank state.
    await expect(page.getByText(/マッチ得点/)).toBeVisible({ timeout: 10_000 });
  });
});
