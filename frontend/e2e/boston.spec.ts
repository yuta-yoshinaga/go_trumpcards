import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Boston E2E', () => {
  test('shows the interleaved ladder and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/boston');

    // Permanent, not tutorial-only: the misere bids sit BETWEEN the trick bids,
    // which is what the whole auction turns on.
    const ladder = page.getByTestId('boston-ladder');
    await expect(ladder).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(ladder).toContainText('リトル・ミゼール');
    await expect(ladder).toContainText('ピッコリッシモ');
    await expect(ladder).toContainText('ミゼールはトリック宣言の間に挟まります');

    await expect(page.getByTestId('boston-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });

    // The human may open the auction, or the CPUs may already have settled it.
    const pass = page.getByRole('button', { name: 'パス' });
    if (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) {
      await pass.click();
      await waitForLoaded(page);
    }

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const alone = page.getByRole('button', { name: '単独で戦う' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(alone, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/boston');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('boston-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
