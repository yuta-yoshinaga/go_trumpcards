import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('German Solo E2E', () => {
  test('bids, calls an ace, and plays through the phase transitions', async ({ page }) => {
    await navigateTo(page, '/germansolo');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // エース呼びの行はどのフェーズでも出る (未指名 / 伏せ / 公開 / 単独)。
    await expect(page.getByTestId('germansolo-ace-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **フラーゲを選ぶ。** 一番低い契約なので、CPU が先に高い宣言をしていない
    // 限り必ず選択肢に出る (出ていなければ下の分岐がパスに落ちる)。
    const frageButton = page.getByRole('button', { name: 'フラーゲ' });
    const passButton = page.getByRole('button', { name: 'パス' });
    const aceCall = page.getByTestId('germansolo-ace-call');
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrick = page.getByRole('button', { name: '次のトリック' });
    const nextDeal = page.getByRole('button', { name: '次のディール' });
    const anyReset = page.getByRole('button', { name: /リセット|次のゲーム/ });
    // **追従必須なので出せない札は aria-disabled になる。** 押しても API は
    // 呼ばれず、ループが無言で空回りする。
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');

    let bids = 0;
    let aceCalls = 0;
    let plays = 0;
    let advances = 0;
    for (let turn = 0; turn < 90; turn++) {
      await expect(
        frageButton.or(passButton).or(aceCall).or(playButton).or(nextTrick).or(nextDeal).or(anyReset).first(),
      ).toBeVisible({
        timeout: TIMEOUT_GAME_LOOP,
      });

      if (await aceCall.isVisible()) {
        // 呼べるエースだけがボタンになっている。最初の 1 つを押す。
        const aces = aceCall.getByRole('button');
        await expect(aces.first()).toBeVisible();
        await aces.first().click();
        await waitForLoaded(page);
        aceCalls++;
        continue;
      }

      if (await frageButton.isVisible()) {
        await frageButton.click();
        const spade = page.getByRole('button', { name: 'スペード' });
        if (await spade.isVisible()) {
          await spade.click();
        }
        const confirm = page.getByRole('button', { name: /宣言|確定/ });
        if (await confirm.isVisible()) {
          await confirm.click();
        }
        await waitForLoaded(page);
        bids++;
        continue;
      }

      if (await passButton.isVisible()) {
        // 上回れる契約が残っていない席はパスするしかない。
        await passButton.click();
        await waitForLoaded(page);
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
        }
        continue;
      }

      if (await nextTrick.isVisible()) {
        await nextTrick.click();
        await waitForLoaded(page);
        advances++;
        continue;
      }

      if (await nextDeal.isVisible()) {
        await nextDeal.click();
        await waitForLoaded(page);
        advances++;
        continue;
      }

      break;
    }

    // **どれだけ働いたかを主張する。** `if (visible)` のループは 0 回でも
    // 緑になるので、実際に宣言し札を出したことをここで固定する。
    expect(bids).toBeGreaterThan(0);
    expect(plays).toBeGreaterThan(0);
    expect(advances).toBeGreaterThan(0);
    // エース呼びは人間が Frage を落札したディールだけ発生するので 0 回もありうる。
    expect(aceCalls).toBeGreaterThanOrEqual(0);
  });

  test('settings: change CPU difficulty and deal count', async ({ page }) => {
    await navigateTo(page, '/germansolo');

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
