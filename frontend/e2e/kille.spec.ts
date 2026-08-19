import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Kille E2E', () => {
  test('shows the pack, keeps other hands hidden and takes a turn', async ({ page }) => {
    await navigateTo(page, '/kille');

    // Permanent, not tutorial-only: an exchanged Harlequin being the LOWEST card
    // is what everyone gets wrong, and it decides whether to trade at all.
    await expect(page.getByTestId('kille-rules-note')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // A single suit means the denomination ladder is the whole of the ranking.
    const ladder = page.getByTestId('kille-ladder');
    await expect(ladder).toContainText('Harlequin');
    await expect(ladder).toContainText('Cuckoo');
    await expect(ladder).toContainText('Mask');

    await expect(page.getByTestId('kille-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });

    // Either action is legal on the human's turn; the dealer's exchange button
    // reads differently because it swaps with the stock.
    //
    // **The human does not always get a turn.** Reset makes seat 0 the dealer,
    // so the human acts LAST, and a CPU that exchanges into the human's Pig
    // knocks the original holder out before then (internal/domain/Kille.go).
    // An eliminated seat renders no action buttons, so waiting longer would
    // never help -- wait for the turn OR for the round to have moved on (#6071).
    const stand = page.getByRole('button', { name: '交換しない' });
    const roundOver = page.getByTestId('kille-showdown').or(page.getByRole('button', { name: '次のラウンドへ' }));
    await expect(stand.or(roundOver).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    if (await stand.isVisible()) {
      await stand.click();
      await waitForLoaded(page);
    }

    // The turn resolves to a showdown or to another seat's turn -- either way
    // the round advances rather than hanging.
    const showdown = page.getByTestId('kille-showdown');
    const next = page.getByRole('button', { name: '次のラウンドへ' });
    const stillPlaying = page.getByRole('button', { name: '交換しない' });
    expect(
      (await isVisibleWithin(showdown, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(stillPlaying, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/kille');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('kille-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
