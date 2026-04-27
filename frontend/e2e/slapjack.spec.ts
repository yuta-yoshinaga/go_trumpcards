import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Slapjack E2E', () => {
  test('reset → step → tick → reset', async ({ page }) => {
    await navigateTo(page, '/slapjack');

    const stepButton = page.getByRole('button', { name: 'めくる' });
    await expect(stepButton).toBeVisible();

    // Take a few human steps; tick polling drives CPU
    for (let i = 0; i < 3; i++) {
      if (await stepButton.isEnabled()) {
        await stepButton.click();
        await waitForLoaded(page);
      }
    }

    // Slap button should always be present in the footer
    const slapButton = page.getByRole('button', { name: 'スラップ！' });
    await expect(slapButton).toBeVisible();

    // Reset returns the page to the initial state
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(stepButton).toBeVisible();
  });
});
