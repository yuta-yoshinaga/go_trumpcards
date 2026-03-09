import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Hearts E2E', () => {
  test('plays a full game: reset → pass/play through tricks and rounds → end → reset', async ({ page }) => {
    await navigateTo(page, '/hearts');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Verify round/trick info is visible (use regex to match "ラウンド 1" header, not score rows)
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    // Verify hearts status is shown
    const heartsBroken = page.getByText('ハートブレイク済');
    const heartsNotBroken = page.getByText('ハート未ブレイク');
    await expect(heartsBroken.or(heartsNotBroken)).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア')).toBeVisible();

    // Game loop: handle pass phase, play phase, trick end, round end
    const MAX_TURNS = 500;
    const passButton = page.getByRole('button', { name: 'パス' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]');

    let gameEnded = false;

    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Wait for any actionable element to appear
      await expect(
        passButton.or(playButton).or(nextTrickButton).or(nextRoundButton).or(resetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      // Check game end: phase 4 shows no action buttons except reset
      const passVisible = await passButton.isVisible().catch(() => false);
      const playVisible = await playButton.isVisible().catch(() => false);
      const nextTrickVisible = await nextTrickButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);

      if (!passVisible && !playVisible && !nextTrickVisible && !nextRoundVisible) {
        gameEnded = true;
        break;
      }

      // Pass phase: select 3 cards and pass
      if (passVisible) {
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

      // Play phase: select a card and play
      if (playVisible) {
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

      // Trick end: proceed to next trick
      if (nextTrickVisible) {
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end: proceed to next round
      if (nextRoundVisible) {
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    // Assert game ended
    expect(gameEnded).toBe(true);

    // Reset for a new game
    await resetButton.click();
    await waitForLoaded(page);

    // Verify game restarted
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();
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
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
