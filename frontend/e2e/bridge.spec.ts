import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Bridge E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/bridge');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    const bidButton = page.getByRole('button', { name: 'ビッド' });
    const passButton = page.getByRole('button', { name: 'パス' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        bidButton.or(passButton).or(playButton).or(nextTrickButton).or(nextRoundButton).or(resetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const bidVisible = await bidButton.isVisible().catch(() => false);
      const passVisible = await passButton.isVisible().catch(() => false);
      const playVisible = await playButton.isVisible().catch(() => false);
      const nextTrickVisible = await nextTrickButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);

      // Game end: no action buttons visible
      if (!bidVisible && !passVisible && !playVisible && !nextTrickVisible && !nextRoundVisible) break;

      // Bid phase: pass to simplify
      if (passVisible && !playVisible) {
        interactions++;
        await passButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Play phase: select a card and play
      if (playVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible().catch(() => false)) && (await playButton.isEnabled())) {
          await playButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Trick end
      if (nextTrickVisible) {
        interactions++;
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    // Verify we had at least one interaction
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty', async ({ page }) => {
    await navigateTo(page, '/bridge');

    // Open settings
    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    // Change CPU difficulty
    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
