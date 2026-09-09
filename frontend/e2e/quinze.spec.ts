import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Quinze E2E', () => {
  test('bets, deals, plays a turn, and settles', async ({ page }) => {
    await navigateTo(page, '/quinze');

    await expect(page.getByText(/チップ/)).toBeVisible();

    const betButton = page.getByRole('button', { name: '100', exact: true });
    const dealButton = page.getByRole('button', { name: '配る', exact: true });
    if (await betButton.isVisible()) {
      await betButton.click();
    } else {
      await expect(dealButton).toBeVisible();
      await dealButton.click();
    }
    await waitForLoaded(page);

    // The initial cards are random, so use whichever legal control the deal offers.
    const hitButton = page.getByRole('button', { name: '引く', exact: true });
    const standButton = page.getByRole('button', { name: '止める', exact: true });
    const bankerStandButton = page.getByRole('button', { name: '親として止める', exact: true });
    if (await hitButton.isVisible()) {
      await hitButton.click();
      await waitForLoaded(page);
      if (await standButton.isVisible()) {
        await standButton.click();
      }
    } else if (await standButton.isVisible()) {
      await standButton.click();
    } else {
      await expect(bankerStandButton).toBeVisible();
      await bankerStandButton.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText('精算', { exact: true })).toBeVisible();
  });
});
