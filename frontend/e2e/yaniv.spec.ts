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

    const MAX_TURNS = 100;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      if (await nextRound.isEnabled()) {
        interactions++;
        await nextRound.click();
        await waitForLoaded(page);
        continue;
      }
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
        // Discard phase: select a card, then discard.
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
      if (
        !(await drawStock.isEnabled()) &&
        !(await nextRound.isEnabled()) &&
        !(await yanivBtn.isEnabled()) &&
        !(await discardBtn.isEnabled())
      ) {
        break;
      }
    }

    expect(interactions).toBeGreaterThan(0);
  });
});
