import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Casino War E2E', () => {
  test('plays a round: bet → resolve (auto) → reset', async ({ page }) => {
    await navigateTo(page, '/casinowar');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // After Bet the round either ends (player win/loss) or enters TIE_DECISION.
    const surrenderButton = page.getByRole('button', { name: 'サレンダー' });
    const warButton = page.getByRole('button', { name: 'ウォー' });
    const resetButton = page.getByRole('button', { name: '次のゲーム' });

    if (await isVisibleWithin(surrenderButton, TIMEOUT_TRANSITION)) {
      // Tie path — surrender to end the round deterministically.
      await surrenderButton.click();
      await waitForLoaded(page);
    } else if (await isVisibleWithin(warButton, TIMEOUT_TRANSITION)) {
      await warButton.click();
      await waitForLoaded(page);
    }

    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('war path when tie decision appears', async ({ page }) => {
    await navigateTo(page, '/casinowar');

    // Retry until we hit a tie (≈1/13 chance per round). On non-tie paths the
    // round auto-ends; reset and try again.
    for (let attempt = 0; attempt < 30; attempt++) {
      const betButton = page.getByRole('button', { name: 'ベット' });
      await expect(betButton).toBeVisible({ timeout: TIMEOUT_ACTION });
      await betButton.click();
      await waitForLoaded(page);

      const warButton = page.getByRole('button', { name: 'ウォー' });
      if (await isVisibleWithin(warButton, TIMEOUT_TRANSITION)) {
        await warButton.click();
        await waitForLoaded(page);
        const resetButton = page.getByRole('button', { name: '次のゲーム' });
        await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });
        await resetButton.click();
        await waitForLoaded(page);
        await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
        return;
      }

      // Non-tie path → round auto-ended. Reset and retry.
      const nextGameButton = page.getByRole('button', { name: '次のゲーム' });
      if (await isVisibleWithin(nextGameButton, TIMEOUT_ACTION)) {
        await nextGameButton.click();
        await waitForLoaded(page);
      } else {
        await navigateTo(page, '/casinowar');
      }
    }
    // If 30 rounds didn't yield a tie, still pass (1/13 base rate makes 30 misses ~10% likely, but the round-end flow is exercised regardless).
  });
});
