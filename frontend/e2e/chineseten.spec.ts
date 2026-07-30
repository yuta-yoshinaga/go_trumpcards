import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Chinese Ten E2E', () => {
  test('plays a card and shows both seats captures', async ({ page }) => {
    await navigateTo(page, '/chineseten');

    // Both capture rules are permanent, not tutorial-only.
    await expect(page.getByText(/A〜9は合計10で取る/)).toBeVisible();
    await expect(page.getByText(/あなたの取り札/)).toBeVisible();
    await expect(page.getByText(/CPU の取り札/)).toBeVisible();

    // Scope to the hand container: `[data-hint-action="play"]` is unique here,
    // but scoping keeps it stable if the page grows another such control.
    const handCards = page.locator('[data-tutorial="ct-hand"] button[data-hint-action="play"]');
    await expect(handCards.first()).toBeVisible();
    const before = await handCards.count();

    await handCards.first().click();
    await waitForLoaded(page);

    // The CPU answers inside the same request, so the hand has shrunk by the
    // time control returns. No assertion on card values -- the deal is shuffled.
    await expect(handCards).toHaveCount(before - 1);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/chineseten');

    const handCards = page.locator('[data-tutorial="ct-hand"] button[data-hint-action="play"]');
    await handCards.first().click();
    await waitForLoaded(page);

    // Mid-game the reset button opens a confirm dialog, whose confirm control
    // is labelled 確認 (common.json button.confirm), not リセット.
    await page.getByRole('button', { name: 'リセット' }).click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(handCards).toHaveCount(12);
  });
});
