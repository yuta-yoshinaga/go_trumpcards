import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('War (戦争) E2E', () => {
  test('plays several rounds: step → resolve → reset', async ({ page }) => {
    await navigateTo(page, '/war');

    // Step button should be visible after initial reset
    const stepButton = page.getByRole('button', { name: 'めくる' });
    await expect(stepButton).toBeVisible();

    // Advance several steps; each click either reveals, resolves, or handles war
    for (let i = 0; i < 8; i++) {
      if (await stepButton.isEnabled()) {
        await stepButton.click();
        await waitForLoaded(page);
      }
    }

    // Reset
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(stepButton).toBeVisible();
  });
});
