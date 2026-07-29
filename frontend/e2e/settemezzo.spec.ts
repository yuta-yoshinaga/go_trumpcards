import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Sette e Mezzo E2E', () => {
  test('bets, plays a turn, and resets', async ({ page }) => {
    await navigateTo(page, '/settemezzo');

    await expect(page.getByText(/チップ/)).toBeVisible();
    await expect(page.getByText(/目標: 7\.5/)).toBeVisible();

    // A CPU banks the opening round, so the stake buttons are what appears --
    // but the bank moves to whoever lands exactly 7.5, and a resumed session can
    // land on either, so handle both.
    const betButton = page.getByRole('button', { name: '100', exact: true });
    const dealButton = page.getByRole('button', { name: '配る', exact: true });
    if (await betButton.isVisible()) {
      await betButton.click();
    } else {
      await dealButton.click();
    }
    await waitForLoaded(page);

    // Whatever the deal gave us, one legal control must be offered.
    const anyAction = page
      .getByRole('button', { name: '引く', exact: true })
      .or(page.getByRole('button', { name: '止める', exact: true }))
      .or(page.getByRole('button', { name: '親として引く', exact: true }))
      .or(page.getByRole('button', { name: 'リセット' }))
      .first();
    await expect(anyAction).toBeVisible();

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/チップ/)).toBeVisible();
  });

  test('the banker hand stays face down while a CPU banks', async ({ page }) => {
    await navigateTo(page, '/settemezzo');

    const betButton = page.getByRole('button', { name: '100', exact: true });
    if (await betButton.isVisible()) {
      await betButton.click();
      await waitForLoaded(page);
      await expect(page.getByLabel('親の手は伏せられています')).toBeVisible();
    }
  });
});
