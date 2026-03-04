import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('BlackJack E2E', () => {
  test('plays a full round: bet → stand → result → next round', async ({ page }) => {
    await navigateTo(page, '/');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION or INSURANCE phase should appear after bet
    // If insurance is offered, decline it
    const declineButton = page.getByRole('button', { name: '辞退' });
    if (await declineButton.isVisible({ timeout: 2_000 }).catch(() => false)) {
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

    // Start next round — reset deals a new hand immediately (ACTION or INSURANCE phase)
    await resetButton.click();
    await waitForLoaded(page);

    // New round started: either ACTION controls (スタンド/ヒット) or INSURANCE (辞退) visible
    const newRoundIndicator = page
      .getByRole('button', { name: 'スタンド' })
      .or(page.getByRole('button', { name: '辞退' }));
    await expect(newRoundIndicator.first()).toBeVisible({ timeout: 10_000 });
  });
});
