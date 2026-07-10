import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Beggar-My-Neighbour (ビガー・マイ・ネイバー) E2E', () => {
  test('plays several turns: step → phases → reset', async ({ page }) => {
    await navigateTo(page, '/beggarmyneighbour');

    // Step button should be visible after initial reset
    const stepButton = page.getByRole('button', { name: 'カードを出す' });
    await expect(stepButton).toBeVisible();

    // Advance several steps through various phases
    for (let i = 0; i < 12; i++) {
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
