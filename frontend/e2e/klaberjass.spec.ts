import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Klaberjass E2E', () => {
  test('shows the trump order, bids and reaches play', async ({ page }) => {
    await navigateTo(page, '/klaberjass');

    // Permanent, not tutorial-only: in trumps the jack and nine jump above the
    // ace, which is what decides whether a hand is worth bidding at all.
    const ladder = page.getByTestId('klaberjass-ladder');
    await expect(ladder).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(ladder).toContainText('J (20)');
    await expect(ladder).toContainText('9 (14)');

    await expect(page.getByTestId('klaberjass-player')).toHaveCount(2, { timeout: TIMEOUT_GAME_LOOP });

    // The human may open the bidding, or the CPU may already have taken it.
    const take = page.getByRole('button', { name: '取る' });
    if (await isVisibleWithin(take, TIMEOUT_GAME_LOOP)) {
      await take.click();
      await waitForLoaded(page);
    }

    // Either we are in play, still bidding, or the CPU took it — the deal must
    // progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const pass = page.getByRole('button', { name: 'パス' });
    const next = page.getByRole('button', { name: '次のディールへ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/klaberjass');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('klaberjass-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
