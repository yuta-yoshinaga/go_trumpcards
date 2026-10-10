import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Nap E2E', () => {
  test('loads, resets, and renders the bid/play UI', async ({ page }) => {
    await navigateTo(page, '/nap');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round / trick info renders.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The game starts in the Bid phase; some interactive control must be present
    // (a bid button, the human's play control, or a trick/round advance).
    const actions = page.getByTestId('game-footer-actions');
    const anyControl = actions
      .getByTestId('bid-0')
      .or(actions.getByRole('button', { name: '出す' }))
      .or(actions.getByRole('button', { name: '次のトリック' }))
      .or(actions.getByRole('button', { name: '次のラウンド' }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // If a bid button is offered, place a Pass to progress the bidding round.
    const passButton = actions.getByTestId('bid-0');
    if (await isVisibleWithin(passButton, TIMEOUT_TRANSITION)) {
      await passButton.click({ timeout: TIMEOUT_TRANSITION }).catch(() => {});
      await waitForLoaded(page);
    }

    // Round info still renders after the interaction.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
