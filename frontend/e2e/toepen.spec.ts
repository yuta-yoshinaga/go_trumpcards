import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Toepen E2E', () => {
  test('plays a card and keeps the ranking on screen', async ({ page }) => {
    await navigateTo(page, '/toepen');

    // The inverted ranking is permanent, not tutorial-only.
    await expect(page.getByText(/10 > 9 > 8 > 7 > A > K > Q > J/)).toBeVisible();
    await expect(page.getByText(/第1ハンド/)).toBeVisible();

    // Scope to the hand container so the toep/reset controls cannot match, and
    // to the cards that are actually PLAYABLE: following suit is compulsory, so
    // `.first()` on the whole hand can land on a card the page will ignore.
    const handCards = page.locator('[data-tutorial="toepen-hand"] button[data-hint-action="play"]');
    const playable = page.locator(
      '[data-tutorial="toepen-hand"] button[data-hint-action="play"][aria-disabled="false"]',
    );
    await expect(playable.first()).toBeVisible();
    const before = await handCards.count();

    await playable.first().click();
    await waitForLoaded(page);

    // The CPUs answer inside the same request, so the hand has shrunk by the
    // time control returns. No assertion on card values -- the deal is shuffled.
    await expect(handCards).toHaveCount(before - 1);
  });

  test('toep raises the stake', async ({ page }) => {
    await navigateTo(page, '/toepen');

    await expect(page.getByText(/賭け点: 1/)).toBeVisible();
    await page.getByRole('button', { name: /toep/i }).click();
    await waitForLoaded(page);

    // The stake is at least two now; the CPUs may have raised again on top.
    await expect(page.getByText(/賭け点: 1/)).toHaveCount(0);
  });
});
