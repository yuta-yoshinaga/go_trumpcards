import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

test('skip link moves focus to the main content without changing the route', async ({ page }) => {
  await navigateTo(page, '/hearts');

  // RouteAnnouncer may move focus to <main> on load, so tab order is outside this test.
  // The regression in #8014 was pressing Enter while the skip link had focus.
  const skipLink = page.getByRole('link', { name: 'メインコンテンツへスキップ', exact: true });
  await skipLink.focus();
  await expect(skipLink).toBeFocused();
  await page.keyboard.press('Enter');

  await expect(page).toHaveURL(/\/#\/hearts$/);
  await expect(page.locator('main#main-content')).toBeFocused();
  await expect(page.getByText('お探しのページは見つかりません')).not.toBeVisible();
});
