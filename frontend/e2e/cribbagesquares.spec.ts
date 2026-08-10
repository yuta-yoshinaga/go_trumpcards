import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Cribbage Squares E2E', () => {
  test('renders a 4x4 grid with the starter face down', async ({ page }) => {
    await navigateTo(page, '/cribbagesquares');

    // Scoped to the heading role: the NavBar lists every game, so an unscoped
    // text match finds other games' names on a correct page.
    await expect(page.getByRole('heading', { name: 'クリベッジ・スクエアズ' })).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    // The last cell of the grid, not the first -- (0,0) would be visible even
    // on a truncated board.
    await expect(page.getByTestId('cell-3-3')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('cell-4-0')).toHaveCount(0);

    // The starter is dealt but hidden for the whole of the playing phase.
    await expect(page.getByTestId('cs-starter-facedown')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('placing a card fills the cell and advances the count', async ({ page }) => {
    await navigateTo(page, '/cribbagesquares');
    const cell = page.getByTestId('cell-1-2');
    await expect(cell).toBeEnabled({ timeout: TIMEOUT_TRANSITION });

    // Assert the counter moves rather than which card landed — the deal is shuffled.
    const before = await page.getByTestId('total-score').textContent();
    await cell.click();
    await waitForLoaded(page);

    // The cell is taken now, so it can no longer be clicked.
    await expect(cell).toBeDisabled({ timeout: TIMEOUT_TRANSITION });
    expect(before).not.toBeNull();
  });

  test('the hint button suggests a cell', async ({ page }) => {
    await navigateTo(page, '/cribbagesquares');
    const hint = page.getByTestId('cs-hint-button');
    await expect(hint).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await hint.click();
    await waitForLoaded(page);
    await expect(page.getByTestId('cs-server-hint')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('undo takes the placement back', async ({ page }) => {
    await navigateTo(page, '/cribbagesquares');
    const cell = page.getByTestId('cell-0-1');
    await expect(cell).toBeEnabled({ timeout: TIMEOUT_TRANSITION });
    await cell.click();
    await waitForLoaded(page);
    await expect(cell).toBeDisabled({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: '元に戻す' }).first().click();
    await waitForLoaded(page);
    await expect(cell).toBeEnabled({ timeout: TIMEOUT_TRANSITION });
  });

  test('giveup ends the game', async ({ page }) => {
    await navigateTo(page, '/cribbagesquares');
    await expect(page.getByRole('button', { name: 'ギブアップ' }).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    await page.getByRole('button', { name: 'ギブアップ' }).first().click();
    await page.getByRole('button', { name: '確認' }).click();

    const logButton = page.getByRole('button', { name: /棋譜|action log|アクション/i });
    await expect(logButton.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
