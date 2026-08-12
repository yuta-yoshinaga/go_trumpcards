import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Rikken E2E', () => {
  test('bids a contract and reaches the play phase', async ({ page }) => {
    await navigateTo(page, '/rikken');

    // 競りは上へしか積めない。いちばん強い契約なら必ず通る。
    const openMisere = page.getByRole('button', { name: /^オープンミゼール/ });
    await expect(openMisere).toBeVisible();
    await openMisere.click();
    await waitForLoaded(page);

    // オープンミゼールは切り札を決めないので、そのままプレイに入る。
    await expect(page.getByTestId('rikken-contract')).toContainText('オープンミゼール', {
      timeout: TIMEOUT_ACTION,
    });

    // **出せる札だけを掴む。** フォロー義務があるので先頭は押せないことがある。
    const legal = page.locator('[data-testid="rikken-hand"] button:not([aria-disabled="true"])');
    await expect(legal.first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    const before = await page.locator('[data-testid="rikken-hand"] button').count();
    await legal.first().click();
    await waitForLoaded(page);

    await expect(page.locator('[data-testid="rikken-hand"] button')).toHaveCount(before - 1, {
      timeout: TIMEOUT_ACTION,
    });
  });

  test('passing hands the auction on', async ({ page }) => {
    await navigateTo(page, '/rikken');

    const pass = page.getByRole('button', { name: '降りる' });
    await expect(pass).toBeVisible();
    await pass.click();
    await waitForLoaded(page);

    // 降りたあとは競りのボタンが消え、契約が決まる。
    await expect(page.getByRole('button', { name: '降りる' })).toBeHidden({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('rikken-contract')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
