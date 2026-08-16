import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

/** Your hand. **The only signal that always moves** when you act. */
const hand = (page: Page) => page.getByRole('button', { name: /を出す$/ });

// **合法な札だけを選ぶ。** フォロー義務を満たさない札はサーバが拒否する。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

test.describe('Linger Longer E2E', () => {
  test('navigates to lingerlonger and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');

    await expect(page.getByText(/リンガーロンガー|Linger Longer/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('ll-stock')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **取っても得点にならず、補充できるだけ。** 直感と逆なので必ず出す。
  test('always states that a trick only buys you a card', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-rule')).toContainText(/補充|stock/, { timeout: TIMEOUT_TRANSITION });
  });

  // **配る枚数は人数と同じ。** 4 人なら 4 枚ずつで、残り 36 枚が山札。
  test('deals four cards each and leaves the rest as stock', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(page.getByTestId('ll-stock')).toContainText('36', { timeout: TIMEOUT_TRANSITION });
    for (const id of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ll-seat-${id}`)).toContainText('4', { timeout: TIMEOUT_TRANSITION });
    }
    await expect(page.getByTestId('ll-seat-4')).toHaveCount(0);
  });

  // **押せる操作はサーバが受理する。** 拒否されていれば盤面が動かない。
  test('playing a legal card advances the trick', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const stockBefore = await page.getByTestId('ll-stock').textContent();
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    // トリック番号か山札のどちらかは必ず動く（取れば補充、取れなければ次の手番）。
    await expect
      .poll(async () => page.getByTestId('ll-stock').textContent(), { timeout: TIMEOUT_ACTION })
      .not.toBe(stockBefore);
  });

  // **取ったトリックの前後で手札は減らない。** そこがこのゲームの全部。
  //
  // 以前はこれを「補充した席は必ず配り枚数(4枚)」と書いていたが、`[直前に補充]` が
  // 意味するのは「直前のトリックを取った」だけで、それ以前に落としていないことは
  // 保証しない。落とせば補充が無く 1 枚ずつ減るので、`手札1枚 / 獲得1回` という
  // 正しい状態でテストが落ちていた (#5802)。不変条件は枚数の固定値ではなく
  // 「取ったトリックを挟んで枚数が変わらない」こと。
  test('a seat that wins a trick keeps its hand size', async ({ page }) => {
    await navigateTo(page, '/lingerlonger');
    await expect(hand(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const seat0 = page.getByTestId('ll-seat-0');
    const cardsOf = async (): Promise<number | null> => {
      const m = /手札(\d+)枚/.exec((await seat0.textContent()) ?? '');
      return m?.[1] === undefined ? null : Number(m[1]);
    };
    const stockOf = async (): Promise<number | null> => {
      // ラベルは「山札 残り{{n}}枚」。`山札(\\d+)` では一致しない。
      const m = /残り(\d+)枚/.exec((await page.getByTestId('ll-stock').textContent()) ?? '');
      return m?.[1] === undefined ? null : Number(m[1]);
    };

    // **12 では足りない。** 人間が 1 度も取れずに skip で終わる配りが 20 回中 13 回
    // あった (2026-08-16 実測)。skip は「通った」ではなく「何も確かめていない」なので、
    // 取れるまでの手数を増やして実際に検証が走る割合を上げる。
    for (let i = 0; i < 40; i++) {
      if (await page.getByTestId('ll-result').isVisible()) break;

      // **CPU が打ち終わるのを待つ。** 直後に legalCard を見て break していたため、
      // 自分の手番が戻る前にループを抜けてしまい、40 手回しても skip 率が下がらなかった。
      await expect(legalCard(page).first().or(page.getByTestId('ll-result'))).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
      if (await page.getByTestId('ll-result').isVisible()) break;
      if (!(await legalCard(page).first().isVisible())) break;

      const cardsBefore = await cardsOf();
      const stockBefore = await stockOf();
      await legalCard(page).first().click();
      await expect(hand(page).first().or(page.getByTestId('ll-result'))).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });

      const wonThisTrick = ((await seat0.textContent()) ?? '').includes('[直前に補充]');
      if (!wonThisTrick) continue;

      // **山札が空なら取っても補充できない。** その局面まで不変を求めると、
      // ルールどおりに減った手札を異常と呼ぶことになる。さらに `[直前に補充]` は
      // `lastDrawIdx` の表示で、補充が止まると**更新されず残る**ので、山札が尽きた
      // 後はこの印を「今取った」の証拠に使えない。
      //
      // 読めなかったときに素通りさせない。ラベルが変われば黙って判定が壊れるので、
      // 分からないなら止める。
      expect(stockBefore, '山札の枚数が読めない (ラベルが変わった?)').not.toBeNull();
      if (stockBefore !== null && stockBefore <= 0) break;

      expect(cardsBefore).not.toBeNull();
      expect(await cardsOf()).toBe(cardsBefore);
      return;
    }
    test.skip(true, 'この配りでは人間が 40 トリック以内に 1 度も取れなかった');
  });
});
