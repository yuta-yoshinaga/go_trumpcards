import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Botifarra E2E', () => {
  test('declares trump and plays a card', async ({ page }) => {
    await navigateTo(page, '/botifarra');

    // 競りは無く、親が宣言するか相方に委ねるかの二択。
    const declareBtn = page.getByRole('button', { name: /ハート を宣言/ });
    await expect(declareBtn).toBeVisible();
    await declareBtn.click();
    await waitForLoaded(page);

    // 倍付けの場面が来たら見送って進める。
    const passBtn = page.getByRole('button', { name: 'そのまま進む' });
    if (await passBtn.isVisible()) {
      await passBtn.click();
      await waitForLoaded(page);
    }

    await expect(page.getByTestId('botifarra-trump')).toContainText('ハート', { timeout: TIMEOUT_ACTION });

    // **出せる札だけを掴む。** 勝つ義務があるので、先頭の札は押せないことがふつうです。
    const legal = page.locator('[data-testid="botifarra-hand"] button:not([aria-disabled="true"])');
    await expect(legal.first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    const before = await page.locator('[data-testid="botifarra-hand"] button').count();
    await legal.first().click();
    await waitForLoaded(page);

    await expect(page.locator('[data-testid="botifarra-hand"] button')).toHaveCount(before - 1, {
      timeout: TIMEOUT_ACTION,
    });
  });

  test('delegates the declaration to the partner', async ({ page }) => {
    await navigateTo(page, '/botifarra');

    const delegateBtn = page.getByRole('button', { name: '相方に委ねる' });
    await expect(delegateBtn).toBeVisible();
    await delegateBtn.click();
    await waitForLoaded(page);

    // 委ねられた相方は必ず宣言するので、切り札が決まる。
    await expect(page.getByTestId('botifarra-trump')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByRole('button', { name: '相方に委ねる' })).toBeHidden({ timeout: TIMEOUT_ACTION });
  });
});
