import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Casino War E2E', () => {
  // The tie-decision path (≈1/13 per round) is covered by unit tests
  // (CasinoWarPage.test.tsx, casinowarFormatter.test.ts). Looping in E2E to
  // force a tie is unreliable in a 90s budget — keep this spec to the
  // reset → bet → auto-resolve → reset flow.
  test('plays a round: bet → resolve (auto or tie) → reset', async ({ page }) => {
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
});
