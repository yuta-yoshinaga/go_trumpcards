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
        continue;
      }

      if (await soloButton.isVisible()) {
        // ソロは最上位の宣言なので、CPU が先に宣言していても通る。
        await soloButton.click();
        const spade = page.getByRole('button', { name: 'スペード' });
        if (await spade.isVisible()) {
          await spade.click();
        }
        const confirm = page.getByRole('button', { name: /宣言|確定/ });
        if (await confirm.isVisible()) {
          await confirm.click();
        } else if (await passButton.isVisible()) {
          await passButton.click();
        }
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
