import { expect, test } from '@playwright/test';
import {
  isVisibleWithin,
  navigateTo,
  TIMEOUT_ACTION,
  TIMEOUT_GAME_LOOP,
  TIMEOUT_TRANSITION,
  waitForLoaded,
} from './helpers';

test.describe('Spanish 21 E2E', () => {
  test('plays a round: bet → stand → result', async ({ page }) => {
    // Spanish 21 reaches the action phase on all but about 8.3% of deals.
    // Five independent deals leave roughly a 1-in-250,000 chance of missing
    // the action phase by chance, so exhausting this loop indicates a bug.
    const maxDeals = 5;
    const standButton = page.getByRole('button', { name: 'スタンド' });
    let reachedAction = false;

    await navigateTo(page, '/spanish21');
    // Page heading is rendered (Japanese label by default).
    await expect(page.getByText(/スパニッシュ21|Spanish 21/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });

    for (let deal = 0; deal < maxDeals; deal++) {
      if (deal > 0) {
        // HashRouter navigation does not remount the page, so only a reload
        // runs the mount-time reset and deals a new hand.
        await page.reload();
        await waitForLoaded(page);
      }

      // BET phase: the first wait after each deal gets the full game-loop
      // budget because the CPU may still be resolving the opening deal.
      const betButton = page.getByRole('button', { name: 'ベット' });
      await expect(betButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
      await betButton.click();
      await waitForLoaded(page);

      // ACTION or INSURANCE phase should appear after bet.
      const declineButton = page.getByRole('button', { name: '辞退' });
      if (await isVisibleWithin(declineButton, TIMEOUT_ACTION)) {
        await declineButton.click();
        await waitForLoaded(page);
      }

      if (!(await isVisibleWithin(standButton, TIMEOUT_ACTION))) continue;

      reachedAction = true;
      await standButton.click();
      await waitForLoaded(page);
      break;
    }

    await expect(
      reachedAction,
      `${maxDeals} 回の配りすべてでアクションフェーズに着かなかった (配りの運ではなく、アクションフェーズに到達できない不具合)。never reached the action phase, so the test could not exercise the round`,
    ).toBe(true);

    // END phase: 次のゲーム button should be visible
    // The dealer still plays out after the stand, so this keeps the game-loop
    // budget the original 10s literal had -- TIMEOUT_ACTION would be a tightening.
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
