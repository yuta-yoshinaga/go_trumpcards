import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Hearts E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/hearts');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    // Verify hearts status is shown
    const heartsBroken = page.getByText('ハートブレイク済');
    const heartsNotBroken = page.getByText('ハート未ブレイク');
    await expect(heartsBroken.or(heartsNotBroken)).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア')).toBeVisible();

    const passButton = page.getByRole('button', { name: 'パス' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 60;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        passButton.or(playButton).or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const passVisible = await passButton.isVisible();
      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Game end: no action buttons visible
      if (!passVisible && !playVisible && !nextTrickVisible && !nextRoundVisible) break;

      if (passVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount >= 3) {
          await handCards.nth(0).click();
          await handCards.nth(1).click();
          await handCards.nth(2).click();
        }
        if (await passButton.isEnabled()) {
          await passButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

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

      // Trick end: human played the 4th card (only visible when human completes a trick)
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

    // Verify we had at least one interaction
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts. Button could be either mid-game (リセット) or end (次のゲーム).
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

  test('settings: change CPU difficulty and point limit', async ({ page }) => {
    await navigateTo(page, '/hearts');

    // Open settings
    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    // Change CPU difficulty
    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // Change point limit
    const limitSelect = page.locator('select').nth(1);
    await expect(limitSelect).toBeVisible();
    await limitSelect.selectOption('50');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
