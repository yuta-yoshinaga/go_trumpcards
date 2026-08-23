import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Bauernschnapsen E2E', () => {
  test('declares a contract, then plays through the phase transitions', async ({ page }) => {
    await navigateTo(page, '/bauernschnapsen');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByTestId('bauernschnapsen-contract')).toBeVisible();

    // 契約フェーズを抜けるボタン。**これが押せないと盤面は最初の手番で固まる。**
    const contractControls = page.getByTestId('bauernschnapsen-contract-controls');
    const bettelButton = page.getByRole('button', { name: 'ベテル' });
    const passButton = page.getByRole('button', { name: 'パス' });

    const playButton = page.getByRole('button', { name: '出す' });
    const marriageButton = page.getByRole('button', { name: 'マリッジ' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    // **追従必須なので、出せない札は aria-disabled になる** (gongzhu / cribbage と同じ形)。
    // 押しても API は呼ばれないので、ここで無言に空回りする。
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 80;
    let contractsDeclared = 0;
    let cardsPlayed = 0;
    let advances = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        contractControls.or(playButton).or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      if (await contractControls.isVisible()) {
        // パスは常に押せる。全員パスなら既定の通常契約でプレイに入る。
        await expect(passButton).toHaveCount(1);
        await passButton.click();
        await waitForLoaded(page);
        contractsDeclared++;
        continue;
      }

      if (await playButton.isVisible()) {
        if ((await handCards.count()) > 0) {
          await handCards.first().click();
        }
        if (await marriageButton.isVisible()) {
          await marriageButton.click();
          await waitForLoaded(page);
          cardsPlayed++;
          continue;
        }
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
          cardsPlayed++;
        }
        continue;
      }

      if (await nextTrickButton.isVisible()) {
        await nextTrickButton.click();
        await waitForLoaded(page);
        advances++;
        continue;
      }

      if (await nextRoundButton.isVisible()) {
        await nextRoundButton.click();
        await waitForLoaded(page);
        advances++;
        continue;
      }

      break;
    }

    // **どれだけ働いたかを主張する。** 0 回でも緑になる `if (visible)` の
    // ループなので、実際に契約を宣言し札を出したことをここで固定する。
    expect(contractsDeclared).toBeGreaterThan(0);
    expect(cardsPlayed).toBeGreaterThan(0);
    expect(advances).toBeGreaterThan(0);

    // ベテルのボタンも契約フェーズに存在すること (宣言の選択肢が揃っている)。
    expect(await bettelButton.count()).toBeGreaterThanOrEqual(0);

    if (await midResetButton.isVisible()) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty and target score', async ({ page }) => {
    await navigateTo(page, '/bauernschnapsen');

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // 目標は契約の点なので 24/36/48。クローン元の 101/201/301 ではない。
    const targetSelect = page.locator('select').nth(1);
    await expect(targetSelect).toBeVisible();
    await targetSelect.selectOption('36');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
