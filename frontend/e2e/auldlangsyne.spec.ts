import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test.describe('AuldLangSyne E2E', () => {
  test('navigates to auldlangsyne and renders the initial board', async ({ page }) => {
    await navigateTo(page, '/auldlangsyne');

    // Wait for the move-count readout
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Foundations row. Uses the JA-locale aria-label since the Playwright suite
    // runs against the JA-default browser. Every foundation builds +1, suit
    // ignored, so there is no step suffix.
    await expect(page.getByLabel(/ファンデーション 0 /).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByLabel(/ファンデーション 3 /).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The stock shows deals remaining rather than a face-up next card.
    await expect(page.getByTestId('als-deals-left')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Control buttons while playing
    await expect(page.getByTestId('als-deal-button')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ヒント' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await expect(page.locator('[data-tutorial="als-reset-button"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('dealing consumes the stock', async ({ page }) => {
    await navigateTo(page, '/auldlangsyne');
    await expect(page.getByTestId('als-deals-left')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 48 non-Ace cards, one row dealt at reset -> 11 deals remain. Asserting the
    // count drops (rather than a specific card) keeps this shuffle-independent.
    const before = await page.getByTestId('als-deals-left').textContent();
    await page.getByTestId('als-deal-button').click();
    await expect(page.getByTestId('als-deals-left')).not.toHaveText(before ?? '', { timeout: TIMEOUT_TRANSITION });
  });

  // Selecting a waste is the only way to arm a move, so the aria-pressed flip is
  // the interaction worth pinning. A specific testid is used rather than
  // `.first()` on a card locator: an empty waste renders a disabled button, and
  // clicking that would hang.
  test('selecting a waste arms the move', async ({ page }) => {
    await navigateTo(page, '/auldlangsyne');
    const waste0 = page.getByTestId('als-waste-button-0');
    await expect(waste0).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(waste0).toHaveAttribute('aria-pressed', 'false');

    await waste0.click();
    await expect(waste0).toHaveAttribute('aria-pressed', 'true', { timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/auldlangsyne');
    await expect(page.getByText(/手数/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();

    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
