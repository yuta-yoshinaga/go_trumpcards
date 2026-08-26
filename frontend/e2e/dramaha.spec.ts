import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Dramaha E2E', () => {
  test('shows both halves of the split from the start', async ({ page }) => {
    await navigateTo(page, '/dramaha');
    await waitForLoaded(page);

    // The same five cards play twice — as an Omaha hand and as a draw hand.
    // Both must be on screen, or the player cannot see what they are playing for.
    await expect(page.getByTestId('dramaha-omaha-hand-name')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId('dramaha-draw-hand-name')).toBeVisible();
  });

  test('reaches the draw round and the hand keeps moving after standing pat', async ({ page }) => {
    await navigateTo(page, '/dramaha');
    await waitForLoaded(page);

    const standPat = page.getByTestId('dramaha-draw-standpat');

    // **名前で拾わない。** `getByRole('button', {name: /チェック|コール/})` は
    // このページで 8 個に当たり、その先頭はプレイスタイルの説明バッジ
    // (「Tight Passive：参加少なめでコール中心」) だった。押しても何も起きず、
    // 回帰アサーションが `if (visible)` の中にあったので、テストは
    // **一度も行動ボタンを押さないまま緑**だった。
    // aria-keyshortcuts は c=コール / k=チェックで、行動ボタンにしか付かない。
    const act = page.locator('button[aria-keyshortcuts="c"], button[aria-keyshortcuts="k"]');

    // 曖昧なロケータに戻ったら、ここで気づけるようにする。
    await expect(act, 'action button locator must resolve to exactly one control').toHaveCount(1);

    for (let i = 0; i < 12; i++) {
      if (await standPat.isVisible().catch(() => false)) break;
      try {
        await act.first().waitFor({ state: 'visible', timeout: 10_000 });
      } catch {
        break; // stand-pat も行動ボタンも出ない = 下の到達アサーションが報告する
      }
      await act.first().click();
      await waitForLoaded(page);
    }

    // **到達自体を先に assert する。** 回帰はこの後のクリックにしか無いので、
    // `if (visible)` で包むと、ドローラウンドに着かなかった回は何も検査せず
    // 通ってしまう。
    await expect(standPat, 'never reached the draw round, so the stall regression was never exercised').toBeVisible({
      timeout: 10_000,
    });

    await standPat.click();
    await waitForLoaded(page);
    // **The hand must not stall here.** Before the fix, nothing drove the CPUs
    // after a draw, so the turn sat on a CPU seat and the player was told
    // "it is not your turn" forever.
    await expect(standPat).toBeHidden({ timeout: 10_000 });
  });
});
