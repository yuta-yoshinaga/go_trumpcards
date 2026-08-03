import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Sheng Ji E2E', () => {
  test('explains the trump group and who scores, then progresses', async ({ page }) => {
    await navigateTo(page, '/shengji');

    // Permanent, not tutorial-only: the trump group and who collects the points.
    const trumpNote = page.getByTestId('shengji-trump-note');
    await expect(trumpNote).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(trumpNote).toContainText('切札は切札スートだけではありません');
    await expect(page.getByTestId('shengji-points-note')).toContainText('点を集めるのは守備側');

    await expect(page.getByTestId('shengji-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });

    // The hand opens on the declaring phase, where passing is always available.
    const pass = page.getByTestId('shengji-pass');
    if (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) {
      await expect(page.getByTestId('shengji-declare-rules')).toContainText('強い宣言だけが上書き');
      await pass.click();
      await waitForLoaded(page);
    }

    // The game must progress rather than hang: burying, playing or the next hand.
    const bury = page.getByRole('button', { name: /底牌に埋める/ });
    const play = page.getByRole('button', { name: '出す' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(bury, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/shengji');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('shengji-trump-note')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
