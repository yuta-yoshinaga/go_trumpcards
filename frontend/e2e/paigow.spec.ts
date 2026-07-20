import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Pai Gow Poker E2E', () => {
  test('plays a round: bet → set hands → result → reset', async ({ page }) => {
    await navigateTo(page, '/paigow');

    // BET phase: click ベット (exact match to avoid matching the ChipBetInput ± steppers)
    const betButton = page.getByRole('button', { name: 'ベット', exact: true });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // SET HANDS phase: select 2 cards and click セット
    // Cards should be visible; click on the first two to select them for low hand.
    // Cards are labelled with their suit/rank (cardAlt), so select by the toggle-state attribute.
    const cards = page.locator('[data-tutorial="pg-set-hands"] button[aria-pressed]');
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });
    await cards.nth(0).click();
    await cards.nth(1).click();

    const setButton = page.getByRole('button', { name: 'セット' });
    await expect(setButton).toBeVisible();
    await setButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット', exact: true })).toBeVisible();
  });
});
