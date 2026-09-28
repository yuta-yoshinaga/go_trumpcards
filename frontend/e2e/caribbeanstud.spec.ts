import { expect, test } from '@playwright/test';
import { gameButton, navigateTo, waitForLoaded } from './helpers';

test.describe('Caribbean Stud Poker E2E', () => {
  test('plays a round: bet → call → result → reset', async ({ page }) => {
    await navigateTo(page, '/caribbeanstud');

    // BET phase: click ベット
    const betButton = gameButton(page, 'ベット');
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // ACTION phase: click コール
    const callButton = gameButton(page, 'コール');
    await expect(callButton).toBeVisible({ timeout: 10_000 });
    await callButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });

  test('fold flow: bet → fold → result → reset', async ({ page }) => {
    await navigateTo(page, '/caribbeanstud');

    const betButton = gameButton(page, 'ベット');
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
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });
});
