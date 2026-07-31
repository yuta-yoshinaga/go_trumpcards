import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Kaiser E2E', () => {
  test('shows the scoring cards, bids and reaches play', async ({ page }) => {
    await navigateTo(page, '/kaiser');

    // Permanent, not tutorial-only: the ♥5 and ♠3 together weigh as much as
    // all eight tricks, and the pack is 34 cards rather than the usual 32.
    const specials = page.getByTestId('kaiser-specials');
    await expect(specials).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(specials).toContainText('♥5');
    await expect(specials).toContainText('♠3');

    await expect(page.getByTestId('kaiser-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });

    // The human may open the bidding, or the CPUs may already have taken it.
    const bidSeven = page.getByRole('button', { name: '7 を宣言' });
    if (await isVisibleWithin(bidSeven, TIMEOUT_GAME_LOOP)) {
      // Six is below the floor, so no such button should ever exist.
      await expect(page.getByRole('button', { name: '6 を宣言' })).toHaveCount(0);
      await page.getByRole('button', { name: 'パス' }).click();
      await waitForLoaded(page);
    }

    // The hand must progress rather than hang: either we play, we name trump /
    // discard as declarer, or the hand is already settled.
    const play = page.getByRole('button', { name: '出す' });
    const discard = page.getByRole('button', { name: '捨てる' });
    const next = page.getByRole('button', { name: '次の局へ' });
    const trump = page.getByRole('button', { name: /を切札に/ });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(discard, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(trump, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/kaiser');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('kaiser-specials')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
