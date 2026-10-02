import { expect, type Page, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

function visibleMoveCount(page: Page) {
  return page.getByTestId('phase-indicator').locator(':scope > span').filter({ hasText: /手数/ });
}

test.describe('Beleaguered Castle E2E', () => {
  test('navigates, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/beleagueredcastle');

    await expect(visibleMoveCount(page)).toBeVisible();

    const hintButton = page.getByRole('button', { name: 'ヒント' });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(visibleMoveCount(page)).toBeVisible();
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/beleagueredcastle');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント' })).not.toBeVisible();
  });
});
