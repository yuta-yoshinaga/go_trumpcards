import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Memory E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/memory');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify score table is visible
    await expect(page.getByRole('status', { name: 'スコア' })).toBeVisible();

    // Verify board cards are present (card back images in buttons)
    const boardButtons = page.locator('[data-testid^="board-"]');
    await expect(boardButtons.first()).toBeVisible();

    const nextButton = page.getByRole('button', { name: '次へ' });

    // Play through several turns to verify phase transitions
    const MAX_TURNS = 60;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Wait for something actionable
      await expect(nextButton.or(resetButton).first()).toBeVisible({ timeout: 10_000 });

      const nextVisible = await nextButton.isVisible();

      if (nextVisible) {
        await nextButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Try to flip a card (human turn in flip1 phase)
      const enabledCards = page.locator('[data-testid^="board-"]:not([disabled]):not(.hidden)');
      try {
        await enabledCards.first().click({ timeout: TIMEOUT_ACTION });
      } catch {
        break; // No enabled cards — game ended or CPU turn
      }
      await waitForLoaded(page);

      // If still in flip phase, flip a second card
      try {
        await enabledCards.first().click({ timeout: TIMEOUT_ACTION });
        await waitForLoaded(page);
      } catch {
        // No more enabled cards to flip — continue to next turn
      }
    }

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('status', { name: 'スコア' })).toBeVisible();
  });

  test('settings: change CPU difficulty', async ({ page }) => {
    await navigateTo(page, '/memory');

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
    await expect(page.getByRole('status', { name: 'スコア' })).toBeVisible();
  });
});
