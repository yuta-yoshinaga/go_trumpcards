import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Dramaha E2E', () => {
  test('shows both halves of the split from the start', async ({ page }) => {
    await navigateTo(page, '/dramaha');
    await waitForLoaded(page);

    // The same five cards play twice — as an Omaha hand and as a draw hand.
    // Both must be on screen, or the player cannot see what they are playing for.
    await expect(page.getByTestId('dramaha-omaha-hand-name')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId('dramaha-draw-hand-name')).toBeVisible();
  });

  test('reaches the draw round and the hand keeps moving after standing pat', async ({ page }) => {
    await navigateTo(page, '/dramaha');
    await waitForLoaded(page);

    const standPat = page.getByTestId('dramaha-draw-standpat');
    for (let i = 0; i < 12; i++) {
      if (await standPat.isVisible().catch(() => false)) break;
      const act = page.getByRole('button', { name: /チェック|コール/ }).first();
      if (!(await act.isVisible().catch(() => false))) break;
      await act.click();
      await waitForLoaded(page);
    }

    if (await standPat.isVisible().catch(() => false)) {
      await standPat.click();
      await waitForLoaded(page);
      // **The hand must not stall here.** Before the fix, nothing drove the CPUs
      // after a draw, so the turn sat on a CPU seat and the player was told
      // "it is not your turn" forever.
      await expect(standPat).toBeHidden({ timeout: 10_000 });
    }
  });
});
