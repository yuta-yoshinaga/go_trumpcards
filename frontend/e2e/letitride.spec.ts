import { expect, test } from '@playwright/test';
import { gameButton, navigateTo, waitForLoaded } from './helpers';

test.describe('Let It Ride E2E', () => {
  test('plays a round: bet → let it ride → let it ride → result → reset', async ({ page }) => {
    await navigateTo(page, '/letitride');

    // BET phase: click ベット
    const betButton = gameButton(page, 'ベット');
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // FIRST DECISION phase: click レットイットライド
    const rideButton1 = gameButton(page, 'レットイットライド');
    await expect(rideButton1).toBeVisible({ timeout: 10_000 });
    await rideButton1.click();
    await waitForLoaded(page);

    // SECOND DECISION phase: click レットイットライド
    const rideButton2 = gameButton(page, 'レットイットライド');
    await expect(rideButton2).toBeVisible({ timeout: 10_000 });
    await rideButton2.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });

  test('pull flow: bet → pull → pull → result → reset', async ({ page }) => {
    await navigateTo(page, '/letitride');

    const betButton = gameButton(page, 'ベット');
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // FIRST DECISION: pull (now requires confirming the risk-reduction dialog)
    const pullButton1 = gameButton(page, 'プル');
    await expect(pullButton1).toBeVisible({ timeout: 10_000 });
    await pullButton1.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // SECOND DECISION: pull
    const pullButton2 = gameButton(page, 'プル');
    await expect(pullButton2).toBeVisible({ timeout: 10_000 });
    await pullButton2.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // END phase
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });
});
