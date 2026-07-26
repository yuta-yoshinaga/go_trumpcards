import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Poker E2E', () => {
  test('plays a full round: reset → bet → stand (no exchange) → bet → result → reset', async ({ page }) => {
    await navigateTo(page, '/poker');

    // Click リセット to start a new game (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // DEAL phase: use チェック or コール to proceed
    const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
    const callButton = page.getByRole('button', { name: 'コール', exact: true });
    if (await isVisibleWithin(checkButton, TIMEOUT_ACTION)) {
      await checkButton.click();
    } else {
      await callButton.click();
    }
    await waitForLoaded(page);

    // EXCHANGE phase: click スタンド (no exchange)
    const standButton = page.getByRole('button', { name: 'スタンド' });
    if (await isVisibleWithin(standButton, TIMEOUT_ACTION)) {
      await standButton.click();
      await waitForLoaded(page);
    }

    // SECOND_BET phase: check/call repeatedly until end-state 次のゲーム appears.
    // A single check/call may not conclude the hand if a CPU re-raises.
    //
    // Wait for whichever control appears FIRST rather than probing each one in
    // turn: two sequential 3s probes cost 6s per lap even when the hand has
    // already ended, so a re-raising CPU could push 20 laps past the 90s test
    // timeout. Playwright then tears the context down, and the in-flight
    // waitForSelector reports "Target page, context or browser has been closed"
    // — which reads like a browser crash but is really this timeout (#2443).
    // Racing them returns as soon as any control is actionable, and a genuine
    // stall now fails fast with a clear message instead of burning the budget.
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    for (let i = 0; i < 20; i++) {
      const anyControl = endResetButton.or(checkButton).or(callButton).first();
      await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
      if (await endResetButton.isVisible()) break;
      if (await checkButton.isVisible()) {
        await checkButton.click();
      } else {
        await callButton.click();
      }
      await waitForLoaded(page);
    }

    // END phase: 次のゲーム should be visible
    await expect(endResetButton).toBeVisible({ timeout: 10_000 });

    // Start another round (end state: no confirm dialog)
    await endResetButton.click();
    await waitForLoaded(page);
  });
});
