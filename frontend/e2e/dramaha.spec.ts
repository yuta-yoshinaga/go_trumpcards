import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Dramaha E2E', () => {
  test('shows both halves of the split from the start', async ({ page }) => {
    await navigateTo(page, '/dramaha');
    await waitForLoaded(page);

    // The same five cards play twice — as an Omaha hand and as a draw hand.
    // Both must be on screen, or the player cannot see what they are playing for.
    await expect(page.getByTestId('dramaha-omaha-hand-name')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('dramaha-draw-hand-name')).toBeVisible();
  });

  test('reaches the draw round and the hand keeps moving after standing pat', async ({ page }) => {
    const standPat = page.getByTestId('dramaha-draw-standpat');
    const maxDeals = 6;

    // **名前で拾わない。** `getByRole('button', {name: /チェック|コール/})` は
    // このページで 8 個に当たり、その先頭はプレイスタイルの説明バッジ
    // (「Tight Passive：参加少なめでコール中心」) だった。押しても何も起きず、
    // 回帰アサーションが `if (visible)` の中にあったので、テストは
    // **一度も行動ボタンを押さないまま緑**だった。
    // aria-keyshortcuts は c=コール / k=チェックで、行動ボタンにしか付かない。
    const act = page.locator('button[aria-keyshortcuts="c"], button[aria-keyshortcuts="k"]');

    // 配りによってはドローラウンドの前にハンドが終わる。それは不具合では
    // ないので、ドローラウンドまで着ける配りを引いてから回帰を検査する。
    for (let deal = 0; deal < maxDeals; deal++) {
      if (deal === 0) {
        await navigateTo(page, '/dramaha');
      } else {
        // **同じ URL への navigate では配り直せない。** HashRouter なので再マウント
        // されず、マウント時の reset が走らないまま前の局面が残る (実測で失敗した)。
        await page.reload();
        await waitForLoaded(page);
      }

      // 曖昧なロケータに戻ったら、ここで気づけるようにする。
      await expect(act, 'action button locator must resolve to exactly one control').toHaveCount(1);

      for (let i = 0; i < 12; i++) {
        if (await standPat.isVisible()) break;
        try {
          // 各配り直し直後の最初の待ちは新しいページのゲームループが進むまで
          // 時間がかかるので全体予算を使い、以降は直前の操作後なので短く待つ。
          const budget = i === 0 ? TIMEOUT_GAME_LOOP : TIMEOUT_ACTION;
          await act.first().waitFor({ state: 'visible', timeout: budget });
        } catch {
          break; // stand-pat も行動ボタンも出ない = 次の配りを試す
        }
        await act.first().click();
        await waitForLoaded(page);
      }

      if (await standPat.isVisible()) break;
    }

    // **到達自体を先に assert する。** 回帰はこの後のクリックにしか無いので、
    // `if (visible)` で包むと、ドローラウンドに着かなかった回は何も検査せず
    // 通ってしまう。
    await expect(
      standPat,
      `${maxDeals} 回の配りすべてでドローラウンドに着かなかった (配りの運ではなく、ドローラウンドに到達できない不具合)。never reached the draw round, so the stall regression was never exercised`,
    ).toBeVisible({
      timeout: TIMEOUT_GAME_LOOP,
    });

    await standPat.click();
    await waitForLoaded(page);
    // **The hand must not stall here.** Before the fix, nothing drove the CPUs
    // after a draw, so the turn sat on a CPU seat and the player was told
    // "it is not your turn" forever.
    await expect(standPat).toBeHidden({ timeout: TIMEOUT_GAME_LOOP });
  });
});
