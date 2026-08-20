import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_LOADED, waitForLoaded } from './helpers';

test.describe('Crazy Eights E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/crazyeights');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const drawButton = page.getByRole('button', { name: '引く' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const suitSpade = page.getByRole('button', { name: '♠', exact: true });
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Break cleanly once the game reaches end state (only 次のゲーム remains).
      if (await endResetButton.isVisible()) break;

      // **操作ボタンが 1 つも無い状態は「待てば出る」とは限らない** (#6031)。
      // CPU の手番ではこのページはどのボタンも描画せず、CPU の進行は
      // リクエストの中で完結するので、保留中の更新が無ければ画面はもう変わらない。
      // 出なければ待ち続けずに切り上げる。最後の「1 回は操作した」と
      // 「リセットでラウンドが始まり直す」が、この test の本来の主張。
      // **最初の 1 回だけ予算を長く取る。**ここには盤面の初回描画とリセット直後の
      // 往復が乗るので、負荷の高いランナーでは 10 秒では足りず、1 度も操作しない
      // まま抜けて `interactions > 0` が落ちていた (#6031 の続き。4 回とも
      // Crazy Eights と無関係な PR で、同じシャードの別 spec も一緒に落ちている)。
      const actionAppeared = await isVisibleWithin(
        playButton.or(drawButton).or(nextRoundButton).or(suitSpade).or(endResetButton).first(),
        turn === 0 ? TIMEOUT_LOADED : TIMEOUT_GAME_LOOP,
      );
      if (!actionAppeared) break;

      const playVisible = await playButton.isVisible();
      const drawVisible = await drawButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();
      const suitVisible = await suitSpade.isVisible();

      // Game end: no action buttons visible
      if (!playVisible && !drawVisible && !nextRoundVisible && !suitVisible) break;

      // Suit choice phase: pick spade
      if (suitVisible) {
        await suitSpade.click();
        await waitForLoaded(page);
        interactions++;
        continue;
      }

      // Play phase: select a card and play, or draw
      if (playVisible || drawVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible()) && (await playButton.isEnabled())) {
          await playButton.click();
          await waitForLoaded(page);
          continue;
        }
        // If play not possible, draw
        if ((await drawButton.isVisible()) && (await drawButton.isEnabled())) {
          await drawButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        await nextRoundButton.click();
        await waitForLoaded(page);
        interactions++;
      }
    }

    // Verify we had at least one interaction (play, draw, suit choice, or round end)
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts. Button could be mid-game (リセット) or end (次のゲーム).
    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
