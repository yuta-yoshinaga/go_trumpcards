import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Bid Euchre E2E', () => {
  test('shows the no-kitty note and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/bideuchre');

    // Permanent, not tutorial-only: 24 / 4 = 6 leaves no remainder, so there
    // is no kitty and no turn-up.
    await expect(page.getByTestId('bideuchre-no-kitty')).toContainText('キティはありません', {
      timeout: TIMEOUT_GAME_LOOP,
    });

    await expect(page.getByTestId('bideuchre-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('bideuchre-scores')).toBeVisible();

    // The human may open the auction, or the CPUs may already have settled it.
    const pass = page.getByRole('button', { name: 'パス' });
    if (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) {
      await pass.click();
      await waitForLoaded(page);
    }

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const trump = page.getByRole('button', { name: '指定する' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(trump, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/bideuchre');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('bideuchre-no-kitty')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
