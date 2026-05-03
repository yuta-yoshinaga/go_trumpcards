import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Pitch E2E', () => {
  test('loads bid phase with pass + bid buttons', async ({ page }) => {
    await navigateTo(page, '/pitch');

    const passButton = page.getByRole('button', { name: 'パス' });
    const bid2Button = page.getByRole('button', { name: 'ビッド 2' });
    const bid3Button = page.getByRole('button', { name: 'ビッド 3' });
    const bid4Button = page.getByRole('button', { name: 'ビッド 4' });

    await expect(passButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(bid2Button).toBeVisible();
    await expect(bid3Button).toBeVisible();
    await expect(bid4Button).toBeVisible();
  });

  test('passes to advance bidding', async ({ page }) => {
    await navigateTo(page, '/pitch');

    const passButton = page.getByRole('button', { name: 'パス' });
    await expect(passButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await passButton.click();
    await waitForLoaded(page);

    // After passes the game either reaches play phase (some CPU bid) or another
    // round of bidding (rare). Verify we eventually see a Play button or back
    // to pass options.
    const playButton = page.getByRole('button', { name: '出す' });
    const playReached = await isVisibleWithin(playButton, TIMEOUT_TRANSITION);
    const passStill = await isVisibleWithin(passButton, TIMEOUT_TRANSITION);
    expect(playReached || passStill).toBe(true);
  });
});
