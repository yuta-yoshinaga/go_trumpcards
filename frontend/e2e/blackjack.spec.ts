import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('BlackJack E2E', () => {
  test('plays a round: bet → stand → result', async ({ page }) => {
    await navigateTo(page, '/');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION or INSURANCE phase should appear after bet
    // If insurance is offered, decline it
    const declineButton = page.getByRole('button', { name: '辞退' });
    if (await isVisibleWithin(declineButton, TIMEOUT_ACTION)) {
      await declineButton.click();
      await waitForLoaded(page);
    }

    // ACTION phase: click スタンド
    const standButton = page.getByRole('button', { name: 'スタンド' });
    await expect(standButton).toBeVisible({ timeout: 5_000 });
    await standButton.click();
    await waitForLoaded(page);

    // END phase: リセット button should be visible
    const resetButton = page.getByRole('button', { name: /リセット/ });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });
  });
});
