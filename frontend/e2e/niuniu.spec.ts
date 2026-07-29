import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Niu Niu E2E', () => {
  test('bets, settles, and resets', async ({ page }) => {
    await navigateTo(page, '/niuniu');

    await expect(page.getByText(/チップ/)).toBeVisible();
    // No hand exists before the deal, and the deal settles the round in the
    // same call -- so there is no mid-round state to observe here.
    await expect(page.getByRole('button', { name: '100', exact: true })).toBeVisible();

    await page.getByRole('button', { name: '100', exact: true }).click();
    await waitForLoaded(page);

    // The round settles at the bet, so the stake buttons disappear.
    await expect(page.getByRole('button', { name: '100', exact: true })).toHaveCount(0);
    await expect(page.getByText(/10の倍数/)).toBeVisible();

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '100', exact: true })).toBeVisible();
  });
});
