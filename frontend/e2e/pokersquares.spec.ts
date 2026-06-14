import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Poker Squares E2E', () => {
  test('places all 25 cards and reaches complete phase', async ({ page }) => {
    await navigateTo(page, '/pokersquares');

    // Fill the 5x5 board in row-major order. Each click triggers an API call.
    for (let r = 0; r < 5; r++) {
      for (let c = 0; c < 5; c++) {
        const cell = page.getByTestId(`cell-${r}-${c}`);
        await expect(cell).toBeVisible({ timeout: TIMEOUT_TRANSITION });
        try {
          await cell.click({ timeout: TIMEOUT_ACTION });
        } catch {
          // If the cell is somehow disabled, skip — all 25 should be placed but order may vary on repeat runs.
        }
        await waitForLoaded(page);
      }
    }

    // After 25 placements, the playing-phase buttons should be gone (complete phase).
    const undoGone = !(await isVisibleWithin(page.getByRole('button', { name: '元に戻す' }), TIMEOUT_ACTION));
    const giveUpGone = !(await isVisibleWithin(page.getByRole('button', { name: 'ギブアップ' }), TIMEOUT_ACTION));
    expect(undoGone && giveUpGone).toBeTruthy();

    // Total score element should be present.
    await expect(page.getByTestId('total-score')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/pokersquares');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // After give up, playing buttons should not be visible
    await expect(page.getByRole('button', { name: 'ギブアップ' })).not.toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
