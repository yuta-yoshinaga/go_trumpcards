import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('Accordion E2E', () => {
  test('navigates to accordion and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/accordion');

    // Wait for game state to fully load via a move-count readout guaranteed to render
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/パイル数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Pile 0 should be rendered as a button (aria-label starts with "0:")
    const pile0 = page.getByRole('button', { name: /^0:/ });
    await expect(pile0.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Hint and giveup buttons are visible while playing
    await expect(page.getByRole('button', { name: 'ヒント' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Reset button targeted via data-tutorial to avoid i18n/text coupling
    await expect(page.locator('[data-tutorial="ac-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/accordion');
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();

    // After giveup, the action log button should appear (end phase shows "棋譜")
    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
