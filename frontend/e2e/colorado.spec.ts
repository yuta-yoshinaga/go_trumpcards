import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Colorado E2E', () => {
  test('renders twenty piles and eight foundations', async ({ page }) => {
    await navigateTo(page, '/colorado');

    // Scoped to the heading role: the NavBar lists every game, so an unscoped
    // text match finds other games' names on a correct page.
    await expect(page.getByRole('heading', { name: 'コロラド' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The last of the twenty piles and the last of the eight foundations --
    // asserting the first of each would pass with a truncated board.
    await expect(page.getByTestId('co-tableau-19')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('co-foundation-7')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('drawing moves a card from the stock to the waste', async ({ page }) => {
    await navigateTo(page, '/colorado');
    const draw = page.getByTestId('co-draw-button');
    await expect(draw).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Assert the counter changes rather than a specific card — the deal is shuffled.
    const before = await page.getByText(/山札/).first().textContent();
    await draw.click();
    await waitForLoaded(page);
    await expect(page.getByText(/山札/).first()).not.toHaveText(before ?? '', { timeout: TIMEOUT_TRANSITION });
  });

  test('the waste card can be buried on an arbitrary pile', async ({ page }) => {
    await navigateTo(page, '/colorado');
    const draw = page.getByTestId('co-draw-button');
    await expect(draw).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await draw.click();
    await waitForLoaded(page);

    const waste = page.getByTestId('co-waste-button');
    await waste.click();
    await expect(waste).toHaveAttribute('aria-pressed', 'true', { timeout: TIMEOUT_TRANSITION });

    // Pile 11 holds a card at deal, and Colorado accepts the waste on it
    // whatever the suit or rank — that acceptance is the point of the game.
    const pile = page.getByTestId('co-tableau-11');
    await pile.click();
    await waitForLoaded(page);
    await expect(waste).toHaveAttribute('aria-pressed', 'false', { timeout: TIMEOUT_TRANSITION });
  });

  test('selecting a tableau pile arms the move', async ({ page }) => {
    await navigateTo(page, '/colorado');
    // Pile 0 always holds a card at deal, so it is safe to click without
    // probing (an empty pile renders a disabled button).
    const pile = page.getByTestId('co-tableau-0');
    await expect(pile).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(pile).toHaveAttribute('aria-pressed', 'false');

    await pile.click();
    await expect(pile).toHaveAttribute('aria-pressed', 'true', { timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/colorado');
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();
    await page.getByRole('button', { name: '確認' }).click();

    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
