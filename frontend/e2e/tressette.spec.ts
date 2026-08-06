import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Tressette E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/tressette');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア')).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    // マストフォローに反する札は aria-disabled で選択不可になる (#4718)。
    // 全札から first() を取ると制限札を掴んでクリックが無反応になるため、
    // 合法な札だけに絞る。
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 60;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(playButton.or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first()).toBeVisible({
        timeout: 10_000,
      });

      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      if (!playVisible && !nextTrickVisible && !nextRoundVisible) break;

      if (playVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      if (nextTrickVisible) {
        interactions++;
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts.
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

  test('settings: change CPU difficulty and target points', async ({ page }) => {
    await navigateTo(page, '/tressette');

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    const targetSelect = page.locator('select').nth(1);
    await expect(targetSelect).toBeVisible();
    await targetSelect.selectOption('31');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
