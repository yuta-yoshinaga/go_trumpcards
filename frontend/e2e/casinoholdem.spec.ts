import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe("Casino Hold'em E2E", () => {
  test('plays a round: bet → call → result → reset', async ({ page }) => {
    await navigateTo(page, '/casinoholdem');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // FLOP phase: click コール (2x ante)
    const callButton = page.getByRole('button', { name: /コール/ });
    await expect(callButton).toBeVisible({ timeout: 10_000 });
    await callButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('fold flow: bet → fold → result → reset', async ({ page }) => {
    await navigateTo(page, '/casinoholdem');

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
