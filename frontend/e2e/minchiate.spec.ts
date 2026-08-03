import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Minchiate E2E', () => {
  test('loads, resets, and renders the play UI', async ({ page }) => {
    await navigateTo(page, '/minchiate');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The 40-trump ladder is the one thing a player cannot infer from the cards,
    // so it must be on screen in every phase.
    await expect(page.getByTestId('minchiate-trump-count-note')).toContainText('切札は40枚', {
      timeout: TIMEOUT_TRANSITION,
    });

    // This is a team game — the score readout must be per team, not per player.
    await expect(page.getByTestId('minchiate-team-scores')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Some interactive control must be present: the scarto, the play control,
    // a trick/round advance, or — once the round resolves — reset / next game.
    const anyControl = page
      .getByRole('button', { name: '捨てる' })
      .or(page.getByRole('button', { name: '出す' }))
      .or(page.getByRole('button', { name: '次のトリック' }))
      .or(page.getByRole('button', { name: '次のラウンド' }))
      .or(page.getByRole('button', { name: /リセット|次のゲーム/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // No bidding phase exists; a bid control would mean the page was derived
    // from a bidding sibling without stripping it.
    await expect(page.getByRole('button', { name: /^パス$/ })).toHaveCount(0);

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
