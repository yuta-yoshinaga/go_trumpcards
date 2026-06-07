import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Yaniv E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/yaniv');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    const discardBtn = page.getByTestId('discard-button');
    const yanivBtn = page.getByTestId('yaniv-button');
    const drawStock = page.getByTestId('draw-stock-button');
    const nextRound = page.getByTestId('next-round-button');
    const handCard0 = page.getByTestId('hand-card-0');

    await expect(discardBtn).toBeVisible({ timeout: 10_000 });

    // Yaniv's elimination score (default 200) means a full game far exceeds the
    // 90s test budget, so cap the number of interactions — the test only needs to
    // verify reset and that the human can act through phase transitions.
    // The next-round button is only rendered at round end, so probe it with
    // isVisible() (immediate) rather than isEnabled() (which waits for the
    // element and would hang the whole test when it is absent).
    const MAX_TURNS = 40;
    const TARGET_INTERACTIONS = 10;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS && interactions < TARGET_INTERACTIONS; turn++) {
      if (await nextRound.isVisible()) {
        interactions++;
        await nextRound.click();
        await waitForLoaded(page);
        continue;
      }
      // These controls are always rendered (disabled when not applicable), so
      // isEnabled() resolves immediately.
      if (await yanivBtn.isEnabled()) {
        interactions++;
        await yanivBtn.click();
        await waitForLoaded(page);
        continue;
      }
      if (await drawStock.isEnabled()) {
        interactions++;
        await drawStock.click();
        await waitForLoaded(page);
        continue;
      }
      if (await discardBtn.isVisible()) {
        // Discard phase: the discard button is always rendered but disabled
        // until a card is selected, so pick one first. Cards are rendered (but
        // disabled) at round/game end too, so gate the click on isEnabled().
        if (await handCard0.isEnabled()) {
          await handCard0.click();
        }
        if (await discardBtn.isEnabled()) {
          interactions++;
          await discardBtn.click();
          await waitForLoaded(page);
          continue;
        }
      }
      // Nothing actionable for the human (e.g. game end) — stop.
      break;
    }

    expect(interactions).toBeGreaterThan(0);
  });
});
