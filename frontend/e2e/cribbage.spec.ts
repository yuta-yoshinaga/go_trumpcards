import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Cribbage E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/cribbage');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const discardButton = page.getByRole('button', { name: 'クリブに捨てる' });
    const pegButton = page.getByRole('button', { name: 'カードを出す' });
    const goButton = page.getByRole('button', { name: 'Go' });
    const showNextButton = page.getByRole('button', { name: '次を表示' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        discardButton.or(pegButton).or(goButton).or(showNextButton).or(nextRoundButton).or(resetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const discardVisible = await discardButton.isVisible().catch(() => false);
      const pegVisible = await pegButton.isVisible().catch(() => false);
      const goVisible = await goButton.isVisible().catch(() => false);
      const showNextVisible = await showNextButton.isVisible().catch(() => false);
      const nextRoundVisible = await nextRoundButton.isVisible().catch(() => false);

      // Game end: no action buttons visible
      if (!discardVisible && !pegVisible && !goVisible && !showNextVisible && !nextRoundVisible) break;

      // Discard phase: select 2 cards and discard
      if (discardVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount >= 2) {
          await handCards.nth(0).click();
          await handCards.nth(1).click();
        }
        if (await discardButton.isEnabled()) {
          await discardButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Pegging phase: select a card and play, or go
      if (pegVisible || goVisible) {
        interactions++;
        if (pegVisible) {
          const cardCount = await handCards.count();
          if (cardCount > 0) {
            await handCards.first().click();
          }
          if ((await pegButton.isVisible().catch(() => false)) && (await pegButton.isEnabled())) {
            await pegButton.click();
            await waitForLoaded(page);
            continue;
          }
        }
        // Fall through to go
        if (goVisible || (await goButton.isVisible().catch(() => false))) {
          await goButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Show phase
      if (showNextVisible) {
        interactions++;
        await showNextButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
      }
    }

    // Verify we had at least one interaction
    expect(interactions).toBeGreaterThan(0);

    // Reset and verify game restarts
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
