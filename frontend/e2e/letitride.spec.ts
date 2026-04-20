import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Let It Ride E2E', () => {
  test('plays a round: bet → let it ride → let it ride → result → reset', async ({ page }) => {
    await navigateTo(page, '/letitride');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // FIRST DECISION phase: click レットイットライド
    const rideButton1 = page.getByRole('button', { name: 'レットイットライド' });
    await expect(rideButton1).toBeVisible({ timeout: 10_000 });
    await rideButton1.click();
    await waitForLoaded(page);

    // SECOND DECISION phase: click レットイットライド
    const rideButton2 = page.getByRole('button', { name: 'レットイットライド' });
    await expect(rideButton2).toBeVisible({ timeout: 10_000 });
    await rideButton2.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('pull flow: bet → pull → pull → result → reset', async ({ page }) => {
    await navigateTo(page, '/letitride');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // FIRST DECISION: pull
    const pullButton1 = page.getByRole('button', { name: 'プル' });
    await expect(pullButton1).toBeVisible({ timeout: 10_000 });
    await pullButton1.click();
    await waitForLoaded(page);

    // SECOND DECISION: pull
    const pullButton2 = page.getByRole('button', { name: 'プル' });
    await expect(pullButton2).toBeVisible({ timeout: 10_000 });
    await pullButton2.click();
    await waitForLoaded(page);

    // END phase
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
