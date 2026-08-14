import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Four Seasons E2E', () => {
  test('renders the cross, the corners and the deal base rank', async ({ page }) => {
    await navigateTo(page, '/fourseasons');

    // Scoped to the heading role: the NavBar lists every game, so an unscoped
    // text match finds other games' names on a correct page.
    await expect(page.getByRole('heading', { name: 'フォーシーズンズ' })).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    // The base rank drives every placement rule, so it must be on screen.
    await expect(page.getByTestId('fs-base-rank')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // Five cross piles and four corners.
    await expect(page.getByTestId('fs-tableau-4')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('fs-foundation-3')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('drawing moves a card from the stock to the waste', async ({ page }) => {
    await navigateTo(page, '/fourseasons');
    const draw = page.getByTestId('fs-draw-button');
    await expect(draw).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Assert the counter changes rather than a specific card — the deal is shuffled.
    const before = await page.getByText(/山札/).first().textContent();
    await draw.click();
    await waitForLoaded(page);
    await expect(page.getByText(/山札/).first()).not.toHaveText(before ?? '', { timeout: TIMEOUT_TRANSITION });
  });

  test('selecting a cross pile arms the move', async ({ page }) => {
    await navigateTo(page, '/fourseasons');
    // Pile 0 always holds a card at deal, so it is safe to click without
    // probing (an empty pile renders a disabled button).
    const pile = page.getByTestId('fs-tableau-0');
    await expect(pile).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(pile).toHaveAttribute('aria-pressed', 'false');

    await pile.click();
    await expect(pile).toHaveAttribute('aria-pressed', 'true', { timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/fourseasons');
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();
    await page.getByRole('button', { name: '確認' }).click();

    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
