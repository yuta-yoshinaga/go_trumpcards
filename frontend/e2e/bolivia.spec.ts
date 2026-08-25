import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Bolivia E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/bolivia');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/ラウンド \d+/).first()).toBeVisible();

    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const drawDiscardButton = page.getByRole('button', { name: '捨て札を取る' });
    const meldButton = page.getByRole('button', { name: 'メルドする' });
    const skipMeldButton = page.getByRole('button', { name: 'スキップ' });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const goOutButton = page.getByRole('button', { name: '上がる' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 60;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        drawStockButton
          .or(drawDiscardButton)
          .or(meldButton)
          .or(skipMeldButton)
          .or(discardButton)
          .or(goOutButton)
          .or(nextRoundButton)
          .or(anyResetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      const drawStockVisible = await drawStockButton.isVisible();
      const skipMeldVisible = await skipMeldButton.isVisible();
      const discardVisible = await discardButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Game end
      if (!drawStockVisible && !skipMeldVisible && !discardVisible && !nextRoundVisible) break;

      // Draw phase
      if (drawStockVisible) {
        interactions++;
        await drawStockButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Meld phase: skip (simplest path)
      if (skipMeldVisible) {
        interactions++;
        await skipMeldButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Discard phase
      if (discardVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        await discardButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
        continue;
      }

      break;
    }

    // **1 回動いただけでは通さない。** クローン元の `> 0` は、最初のドローが
    // 効いただけで緑になる ── それでは「フェーズが回る」ことを測っていない。
    // 4 フェーズを一巡すれば必ず 4 回以上になる。
    expect(interactions).toBeGreaterThanOrEqual(4);
  });

  // **このゲームを名指す言葉が画面に出ていること。**
  //
  // 上がりに要るのはエスカレラで、ボリビアは点が重いだけ ── その区別が
  // 画面のどこにも無ければ、遊ぶ側はカナスタ 2 個で上がろうとして詰まる。
  test('names the escalera requirement, not just canastas', async ({ page }) => {
    await navigateTo(page, '/bolivia');
    await expect(page.getByText(/ラウンド \d+/).first()).toBeVisible({ timeout: 15_000 });

    // 上がりの条件がどこかに書いてあること。
    const goOutRule = page.getByText(/エスカレラ/).first();
    await expect(goOutRule).toBeVisible({ timeout: 10_000 });
  });
});
