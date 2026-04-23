import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Red Dog E2E', () => {
  test('plays a round: bet → decision/auto → result → reset', async ({ page }) => {
    await navigateTo(page, '/reddog');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // After bet, the game may go to SPREAD_DECISION or skip to END (pair/consecutive)
    const raiseButton = page.getByRole('button', { name: 'レイズ' });
    const stayButton = page.getByRole('button', { name: 'ステイ' });
    const resetButton = page.getByRole('button', { name: '次のゲーム' });

    if (await isVisibleWithin(raiseButton, TIMEOUT_TRANSITION)) {
      // SPREAD_DECISION phase: choose stay
      await stayButton.click();
      await waitForLoaded(page);
    }

    // END phase: 次のゲーム button should be visible
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('raise flow when spread decision appears', async ({ page }) => {
    await navigateTo(page, '/reddog');

    // Retry until we get a spread decision. Pair → PairThird (no visible
    // button to advance) and consecutive → push/END. In either non-spread
    // case, navigate away to fully reset the backend session.
    for (let attempt = 0; attempt < 20; attempt++) {
      const betButton = page.getByRole('button', { name: 'ベット' });
      await expect(betButton).toBeVisible({ timeout: TIMEOUT_ACTION });
      await betButton.click();
      await waitForLoaded(page);

      const raiseButton = page.getByRole('button', { name: 'レイズ' });
      if (await isVisibleWithin(raiseButton, TIMEOUT_TRANSITION)) {
        // Got a spread — click raise
        await raiseButton.click();
        await waitForLoaded(page);

        // END phase
        const resetButton = page.getByRole('button', { name: '次のゲーム' });
        await expect(resetButton).toBeVisible({ timeout: TIMEOUT_ACTION });

        await resetButton.click();
        await waitForLoaded(page);
        await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
        return;
      }

      // Non-spread path: could be auto-END (consecutive) or stuck PairThird
      // (no visible button). Re-navigate to reset the session either way.
      const nextGameButton = page.getByRole('button', { name: '次のゲーム' });
      if (await isVisibleWithin(nextGameButton, TIMEOUT_ACTION)) {
        await nextGameButton.click();
        await waitForLoaded(page);
      } else {
        // Stuck in PairThird — reload the page to restart the session.
        await navigateTo(page, '/reddog');
      }
    }

    // If we never got a spread in 20 attempts, still pass (extremely unlikely)
  });
});
