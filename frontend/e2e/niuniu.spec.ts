import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Niu Niu E2E', () => {
  test('bets, settles, and starts the next round', async ({ page }) => {
    await navigateTo(page, '/niuniu');

    await expect(page.getByText(/チップ/)).toBeVisible();
    // No hand exists before the deal, and the deal settles the round in the
    // same call -- so there is no mid-round state to observe here.
    const bet = page.getByRole('button', { name: '100', exact: true });
    await expect(bet).toBeVisible();

    await bet.click();
    await waitForLoaded(page);

    // The round settles at the bet: the stake buttons go away and the combo
    // hint appears.
    await expect(bet).toHaveCount(0);
    await expect(page.getByText(/10の倍数/)).toBeVisible();

    // GameResetButton switches to "次のゲーム" once the round has ended, and
    // fires immediately rather than opening the confirm dialog.
    await page.getByRole('button', { name: '次のゲーム' }).click();
    await waitForLoaded(page);

    await expect(bet).toBeVisible();
  });
});
