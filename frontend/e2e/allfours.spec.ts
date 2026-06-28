import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('All Fours E2E', () => {
  test('loads beg phase with stand + beg buttons', async ({ page }) => {
    await navigateTo(page, '/allfours');

    const standButton = page.getByRole('button', { name: 'スタンド' });
    const begButton = page.getByRole('button', { name: 'ベグ' });

    await expect(standButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(begButton).toBeVisible();
  });

  test('stand advances to play phase', async ({ page }) => {
    await navigateTo(page, '/allfours');

    const standButton = page.getByRole('button', { name: 'スタンド' });
    await expect(standButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await standButton.click();
    await waitForLoaded(page);

    // After standing, the human should be able to play (Play button visible) or
    // we are at a trick/deal transition. Verify the game progressed.
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const playReached = await isVisibleWithin(playButton, TIMEOUT_TRANSITION);
    const nextReached = await isVisibleWithin(nextTrickButton, TIMEOUT_TRANSITION);
    expect(playReached || nextReached).toBe(true);
  });
});
