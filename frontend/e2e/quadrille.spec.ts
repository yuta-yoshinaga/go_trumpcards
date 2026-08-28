import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Quadrille E2E', () => {
  test('bids, calls a king, and plays through the phase transitions', async ({ page }) => {
    await navigateTo(page, '/quadrille');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // 王呼びの行はどのフェーズでも出る (未指名 / 伏せ / 公開 / 単独)。
    await expect(page.getByTestId('quadrille-king-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const soloButton = page.getByRole('button', { name: 'ソロ' });
    const passButton = page.getByRole('button', { name: 'パス' });
    const kingCall = page.getByTestId('quadrille-king-call');
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrick = page.getByRole('button', { name: '次のトリック' });
    const nextDeal = page.getByRole('button', { name: '次のディール' });
    const anyReset = page.getByRole('button', { name: /リセット|次のゲーム/ });
    // **追従必須なので出せない札は aria-disabled になる。** 押しても API は
    // 呼ばれず、ループが無言で空回りする。
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');

    let bids = 0;
    let kingCalls = 0;
    let plays = 0;
    let advances = 0;
    // どの枝を何回通ったかを残す。ループが空回りしたとき、`plays === 0` だけ
    // では**どこで止まったのか分からない** (#6689 の調査で実際に困った)。
    const trace: string[] = [];
    for (let turn = 0; turn < 90; turn++) {
      await expect(soloButton.or(kingCall).or(playButton).or(nextTrick).or(nextDeal).or(anyReset).first()).toBeVisible({
        timeout: TIMEOUT_GAME_LOOP,
      });

      if (await kingCall.isVisible()) {
        // 呼べる王だけがボタンになっている。最初の 1 つを押す。
        const kings = kingCall.getByRole('button');
        await expect(kings.first()).toBeVisible();
        await kings.first().click();
        await waitForLoaded(page);
        kingCalls++;
        trace.push('king');
        continue;
      }

      if (await soloButton.isVisible()) {
        await soloButton.click();
        const spade = page.getByRole('button', { name: 'スペード' });
        if (await spade.isVisible()) {
          await spade.click();
        }
        const confirm = page.getByTestId('quadrille-bid-confirm');
        if (await confirm.isVisible()) {
          await confirm.click();
        }
        await waitForLoaded(page);

        // **ソロが常に通るとは限らない。** CPU が先にソロを宣言していると
        // サーバが弾き、画面は宣言の選択 (stage1) に戻る。同じ宣言を押し続けても
        // ループは空回りするだけで、90 手を使い切って `plays === 0` で落ちていた
        // (#6689、実測 1/10)。**パスは常に通る**ので、通らなかったらパスして
        // 手番を進める。
        if (await page.getByTestId('quadrille-bid-stage1').isVisible()) {
          await passButton.click();
          await waitForLoaded(page);
          trace.push('bid:pass');
        } else {
          trace.push('bid:solo');
        }
        bids++;
        continue;
      }

      if (await playButton.isVisible()) {
        if ((await handCards.count()) > 0) {
          await handCards.first().click();
        }
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
          plays++;
          trace.push('play');
        } else {
          trace.push('play:blocked');
        }
        continue;
      }

      if (await nextTrick.isVisible()) {
        await nextTrick.click();
        await waitForLoaded(page);
        advances++;
        trace.push('trick');
        continue;
      }

      if (await nextDeal.isVisible()) {
        await nextDeal.click();
        await waitForLoaded(page);
        advances++;
        trace.push('deal');
        continue;
      }

      break;
    }

    // **どれだけ働いたかを主張する。** `if (visible)` のループは 0 回でも
    // 緑になるので、実際に宣言し札を出したことをここで固定する。
    const where = () => `bids=${bids} kings=${kingCalls} plays=${plays} advances=${advances} | ${trace.join(' ')}`;
    expect(bids, where()).toBeGreaterThan(0);
    expect(plays, where()).toBeGreaterThan(0);
    expect(advances, where()).toBeGreaterThan(0);
    // 王呼びは人間が落札したディールだけ発生するので 0 回もありうる。
    expect(kingCalls).toBeGreaterThanOrEqual(0);
  });

  test('settings: change CPU difficulty and deal count', async ({ page }) => {
    await navigateTo(page, '/quadrille');

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await settingsToggle.click();

    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
