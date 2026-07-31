import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Karnöffel E2E', () => {
  test('shows the irregular ranking and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/karnoffel');

    // Permanent, not tutorial-only: the ranking is the whole point, and the
    // devil's rule cannot be inferred from the cards on screen.
    const ladder = page.getByTestId('karnoffel-ladder');
    await expect(ladder).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(ladder).toContainText('J（カルニッフェル）');
    await expect(ladder).toContainText('7（悪魔・リード時のみ）');
    await expect(page.getByTestId('karnoffel-chosen-note')).toContainText('最も低い札');

    await expect(page.getByTestId('karnoffel-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('karnoffel-scores')).toBeVisible();

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect((await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) || (await isVisibleWithin(next, TIMEOUT_GAME_LOOP))).toBe(
      true,
    );
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/karnoffel');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('karnoffel-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
