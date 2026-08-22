import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Three Card Rummy E2E', () => {
  test('plays a round: bet → play → result → reset', async ({ page }) => {
    await navigateTo(page, '/threecardrummy');

    // BET phase: the scoring note is what tells a first-time player that a LOW
    // total wins, so it has to reach the browser, not just the unit test.
    await expect(page.getByText(/低いほど強く/)).toBeVisible();

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION phase: the player's own total is on screen, the dealer's is not.
    const playButton = page.getByRole('button', { name: 'プレイ' });
    await expect(playButton).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/点数: ？/)).toBeVisible();

    await playButton.click();
    await waitForLoaded(page);

    // END phase: the dealer is revealed and 次のゲーム appears.
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/点数: ？/)).toHaveCount(0);

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('fold flow: bet → fold → result → reset', async ({ page }) => {
    await navigateTo(page, '/threecardrummy');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    const foldButton = page.getByRole('button', { name: 'フォールド' });
    await expect(foldButton).toBeVisible({ timeout: 10_000 });
    await foldButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
