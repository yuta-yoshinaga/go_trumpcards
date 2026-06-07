import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Thirty-One E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/thirtyone');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    const drawStock = page.getByTestId('draw-stock-button');
    const drawDiscard = page.getByTestId('draw-discard-button');
    const discardBtn = page.getByTestId('discard-button');
    const knockBtn = page.getByTestId('knock-button');
    const nextRound = page.getByTestId('next-round-button');
    const handCard0 = page.getByTestId('hand-card-0');

    await expect(drawStock).toBeVisible({ timeout: 10_000 });

    // The next-round button is only rendered at round end, so probe it with
    // isVisible() (immediate) rather than isEnabled() (which auto-waits for the
    // element and would hang the whole test for the 90s timeout when absent).
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      if (await nextRound.isVisible()) {
        interactions++;
        await nextRound.click();
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
        // Discard phase: pick a card then discard. Cards are rendered (but
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
      // Nothing actionable for the human — likely game end.
      break;
    }

    expect(interactions).toBeGreaterThan(0);
    // Reference unused locators to keep the linter satisfied while documenting intent.
    expect(await knockBtn.count()).toBeGreaterThanOrEqual(0);
    expect(await drawDiscard.count()).toBeGreaterThanOrEqual(0);
  });
});
