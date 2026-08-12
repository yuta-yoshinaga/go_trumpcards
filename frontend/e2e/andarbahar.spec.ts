import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Andar Bahar E2E', () => {
  test('plays a round: bet on andar → auto-resolve → next round', async ({ page }) => {
    await navigateTo(page, '/andarbahar');

    // **先に配る列は賭ける前に見えている。** 配当が 0.9:1 に下がる側です。
    await expect(page.getByTestId('andarbahar-first-column')).toBeVisible();

    const andarBtn = page.getByRole('button', { name: 'アンダーに賭ける' });
    await expect(andarBtn).toBeVisible();
    await andarBtn.click();
    await waitForLoaded(page);

    // 決着まで一気に進むので、配った枚数と結果が出る。
    await expect(page.getByTestId('andarbahar-dealt-count')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('payout-result')).toBeVisible({ timeout: TIMEOUT_ACTION });

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'アンダーに賭ける' })).toBeVisible();
  });

  test('places a side bet through the band selector', async ({ page }) => {
    await navigateTo(page, '/andarbahar');

    // **セレクトは設定パネルの外に置いてある。** 閉じた <details> の中だと
    // jsdom では触れて Playwright ではタイムアウトします。
    const bandSelect = page.getByLabel('サイドベット');
    await expect(bandSelect).toBeVisible();
    await bandSelect.selectOption('2');

    await expect(page.getByLabel('サイドベット額')).toBeVisible({ timeout: TIMEOUT_ACTION });

    await page.getByRole('button', { name: 'バハールに賭ける' }).click();
    await waitForLoaded(page);
    await expect(page.getByTestId('payout-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
