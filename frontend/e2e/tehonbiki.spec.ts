import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Tehonbiki E2E', () => {
  test('places a wager and starts the next round', async ({ page }) => {
    await navigateTo(page, '/tehonbiki');
    await expect(page.getByRole('button', { name: '張る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '1' })).toBeVisible();
    await expect(page.getByRole('button', { name: '6' })).toBeVisible();

    await page.getByRole('button', { name: '張る' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '次の勝負' })).toBeVisible({ timeout: TIMEOUT_ACTION });
    await page.getByRole('button', { name: '次の勝負' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '張る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/tehonbiki');
    await expect(page.getByRole('button', { name: '張る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /リセット|Reset/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認|Confirm/ });
    if (await confirm.first().isVisible({ timeout: TIMEOUT_ACTION })) await confirm.first().click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '張る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
