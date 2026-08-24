import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Gleek E2E', () => {
  test('bids, discards, and plays through the phase transitions', async ({ page }) => {
    await navigateTo(page, '/gleek');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // ストックの行はどのフェーズでも出る (競り中 / 落札済み)。
    await expect(page.getByTestId('gleek-stage-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const raiseButton = page.getByRole('button', { name: /まで競り上げる/ });
    const dropButton = page.getByRole('button', { name: '降りる' });
    const discardConfirm = page.getByTestId('gleek-discard-confirm');
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrick = page.getByRole('button', { name: '次のトリック' });
    const nextDeal = page.getByRole('button', { name: '次のディール' });
    const anyReset = page.getByRole('button', { name: /リセット|次のゲーム/ });
    // **出せない札は aria-disabled になる。** 押しても API は呼ばれず、
    // ループが無言で空回りする。
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');

    let bids = 0;
    let discards = 0;
    let plays = 0;
    let advances = 0;
    for (let turn = 0; turn < 120; turn++) {
      await expect(
        raiseButton.or(dropButton).or(discardConfirm).or(playButton).or(nextTrick).or(nextDeal).or(anyReset).first(),
      ).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

      if (await discardConfirm.isVisible()) {
        // 確定できるまで札を選ぶ。ちょうどの枚数でしか有効にならない。
        for (let i = 0; i < 12 && !(await discardConfirm.isEnabled()); i++) {
          const card = handCards.nth(i);
          if (!(await card.isVisible())) break;
          await card.click();
        }
        if (await discardConfirm.isEnabled()) {
          await discardConfirm.click();
          await waitForLoaded(page);
          discards++;
        }
        continue;
      }

      if (await dropButton.isVisible()) {
        // 降りる側を押す。競りは 1 人残れば必ず閉じる。
        await dropButton.click();
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
    // 緑になるので、実際に競って札を出したことをここで固定する。
    expect(bids).toBeGreaterThan(0);
    expect(plays).toBeGreaterThan(0);
    expect(advances).toBeGreaterThan(0);
    // 捨て札は人間が落札したディールだけ発生するので 0 回もありうる。
    expect(discards).toBeGreaterThanOrEqual(0);
  });

  test('settings: change CPU difficulty and deal count', async ({ page }) => {
    await navigateTo(page, '/gleek');

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
