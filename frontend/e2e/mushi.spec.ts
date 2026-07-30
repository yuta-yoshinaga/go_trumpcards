import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Mushi E2E', () => {
  test('plays a card and shows both seats captures', async ({ page }) => {
    await navigateTo(page, '/mushi');

    await expect(page.getByText(/第1局/)).toBeVisible();
    // Captures are public for both seats -- that is the game's core read.
    await expect(page.getByText(/あなたの取り札/)).toBeVisible();
    await expect(page.getByText(/CPU の取り札/)).toBeVisible();

    // Scope to the hand container: `[data-hint-action="play"]` alone would
    // also match nothing else here, but scoping keeps it stable if the page
    // grows another control with that attribute.
    const handCards = page.locator('[data-tutorial="mushi-hand"] button[data-hint-action="play"]');
    await expect(handCards.first()).toBeVisible();
    const before = await handCards.count();

    await handCards.first().click();
    await waitForLoaded(page);

    // The CPU answers inside the same request, so the human's hand has shrunk
    // by the time control returns. No assertion on card values -- the deal is
    // shuffled.
    await expect(handCards).toHaveCount(before - 1);
  });

  test('can be reset', async ({ page }) => {
    await navigateTo(page, '/mushi');

    const handCards = page.locator('[data-tutorial="mushi-hand"] button[data-hint-action="play"]');
    await handCards.first().click();
    await waitForLoaded(page);

    // Mid-game the reset button opens a confirm dialog, whose confirm control
    // is labelled 確認 (common.json button.confirm), not リセット.
    await page.getByRole('button', { name: 'リセット' }).click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/第1局/)).toBeVisible();
  });
});
