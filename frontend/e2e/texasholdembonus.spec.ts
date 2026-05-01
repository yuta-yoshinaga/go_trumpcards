import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe("Texas Hold'em Bonus Poker E2E", () => {
  test('plays a round: bet → play → check → check → result → reset', async ({ page }) => {
    await navigateTo(page, '/texasholdembonus');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // PRE-FLOP phase: click プレイ (2× ante)
    const playButton = page.getByRole('button', { name: /プレイ/ });
    await expect(playButton).toBeVisible({ timeout: 10_000 });
    await playButton.click();
    await waitForLoaded(page);

    // FLOP phase: click チェック
    const checkButton = page.getByRole('button', { name: 'チェック' });
    await expect(checkButton).toBeVisible({ timeout: 10_000 });
    await checkButton.click();
    await waitForLoaded(page);

    // TURN phase: click チェック again
    const checkButton2 = page.getByRole('button', { name: 'チェック' });
    await expect(checkButton2).toBeVisible({ timeout: 10_000 });
    await checkButton2.click();
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
    await navigateTo(page, '/texasholdembonus');

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

  test('raise flow: bet → play → raise → raise → result', async ({ page }) => {
    await navigateTo(page, '/texasholdembonus');

    await page.getByRole('button', { name: 'ベット' }).click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: /プレイ/ }).click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: /レイズ/ }).click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: /レイズ/ }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '次のゲーム' })).toBeVisible({ timeout: 10_000 });
  });
});
