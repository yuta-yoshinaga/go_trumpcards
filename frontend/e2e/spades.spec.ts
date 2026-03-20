import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Spades E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/spades');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    // Verify spades status is shown
    const spadesBroken = page.getByText('スペードブレイク済');
    const spadesNotBroken = page.getByText('スペード未ブレイク');
    await expect(spadesBroken.or(spadesNotBroken)).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア')).toBeVisible();

    const bidButton = page.getByRole('button', { name: 'ビッド' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const bidInput = page.locator('input[aria-label="bid-input"]');
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 60;
    let sawPlay = false;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        bidButton.or(playButton).or(nextTrickButton).or(nextRoundButton).or(resetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const bidVisible = await bidButton.isVisible().catch(() => false);
      const playVisible = await playButton.isVisible().catch(() => false);
      const nextTrickVisible = await nextTrickButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);

      // Game end: no action buttons visible
      if (!bidVisible && !playVisible && !nextTrickVisible && !nextRoundVisible) break;

      // Bid phase: enter a bid
      if (bidVisible) {
        await bidInput.fill('3');
        await bidButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Play phase: select a card and play
      if (playVisible) {
        sawPlay = true;
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

      // Trick end
      if (nextTrickVisible) {
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    // Verify we saw play phase
    expect(sawPlay).toBe(true);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty and point limit', async ({ page }) => {
    await navigateTo(page, '/spades');

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
    await limitSelect.selectOption('300');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
