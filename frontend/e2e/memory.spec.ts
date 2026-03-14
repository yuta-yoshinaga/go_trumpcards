import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Memory E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/memory');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await waitForLoaded(page);

    // Verify score table is visible
    await expect(page.getByText('スコア')).toBeVisible();

    // Verify board cards are present (52 numbered card buttons)
    const boardButtons = page.locator('button').filter({ hasText: /^\d+$/ });
    await expect(boardButtons.first()).toBeVisible();

    const nextButton = page.getByRole('button', { name: '次へ' });

    // Play through several turns to verify phase transitions
    const MAX_TURNS = 60;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      // Wait for something actionable
      await expect(nextButton.or(resetButton).first()).toBeVisible({ timeout: 10_000 });

      const nextVisible = await nextButton.isVisible().catch(() => false);

      if (nextVisible) {
        await nextButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Try to flip a card (human turn in flip1 phase)
      // Find an enabled board card and click it
      const enabledCards = page.locator('button:not([disabled])').filter({ hasText: /^\d+$/ });
      const cardCount = await enabledCards.count();
      if (cardCount === 0) break; // Game may have ended or CPU turn

      await enabledCards.first().click();
      await waitForLoaded(page);

      // If still in flip phase, flip a second card
      const enabledCards2 = page.locator('button:not([disabled])').filter({ hasText: /^\d+$/ });
      const cardCount2 = await enabledCards2.count();
      if (cardCount2 > 0) {
        await enabledCards2.first().click();
        await waitForLoaded(page);
      }
    }

    // Reset and verify game restarts
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByText('スコア')).toBeVisible();
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
    await waitForLoaded(page);

    // Verify game started
    await expect(page.getByText('スコア')).toBeVisible();
  });
});
