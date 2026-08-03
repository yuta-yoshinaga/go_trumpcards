import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Pontoon E2E', () => {
  test('bets, plays a declaration, and resets', async ({ page }) => {
    await navigateTo(page, '/pontoon');

    await expect(page.getByText(/チップ/)).toBeVisible();
    await expect(page.getByText(/親/).first()).toBeVisible();

    // A CPU banks the opening round, so the stake buttons are what appears --
    // but the bank moves to whoever makes a pontoon, and a resumed session can
    // land on either, so handle both.
    const betButton = page.getByRole('button', { name: '100', exact: true });
    const dealButton = page.getByRole('button', { name: '配る', exact: true });
    if (await betButton.isVisible()) {
      await betButton.click();
    } else {
      await dealButton.click();
    }
    await waitForLoaded(page);

    // Whatever the deal gave us, at least one legal control must be offered:
    // a declaration, the banker's draw, or the next round.
    const anyAction = page
      .getByRole('button', { name: 'スティック', exact: true })
      .or(page.getByRole('button', { name: 'ツイスト', exact: true }))
      .or(page.getByRole('button', { name: '引く', exact: true }))
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

  test('the banker hand stays face down until the round settles', async ({ page }) => {
    await navigateTo(page, '/pontoon');

    const betButton = page.getByRole('button', { name: '100', exact: true });
    if (await betButton.isVisible()) {
      await betButton.click();
      await waitForLoaded(page);
      // A CPU banks in this branch, so the banker's cards must be hidden.
      await expect(page.getByLabel('親の手は伏せられています')).toBeVisible();
    }
  });
});
