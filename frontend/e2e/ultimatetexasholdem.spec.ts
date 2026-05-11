import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe("Ultimate Texas Hold'em E2E", () => {
  test('check-through flow: bet → check → check → play 1× → result → reset', async ({ page }) => {
    await navigateTo(page, '/ultimatetexasholdem');

    // BET phase
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // PRE-FLOP: check
    const preFlopCheck = page.getByRole('button', { name: 'チェック' });
    await expect(preFlopCheck).toBeVisible({ timeout: 10_000 });
    await preFlopCheck.click();
    await waitForLoaded(page);

    // FLOP: check again (will deal turn + river and enter river phase)
    const flopCheck = page.getByRole('button', { name: 'チェック' });
    await expect(flopCheck).toBeVisible({ timeout: 10_000 });
    await flopCheck.click();
    await waitForLoaded(page);

    // RIVER: play 1×
    const play1x = page.getByRole('button', { name: 'プレイ 1×' });
    await expect(play1x).toBeVisible({ timeout: 10_000 });
    await play1x.click();
    await waitForLoaded(page);

    // END: reset
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('preflop 4× flow: bet → play 4× → result', async ({ page }) => {
    await navigateTo(page, '/ultimatetexasholdem');

    await page.getByRole('button', { name: 'ベット' }).click();
    await waitForLoaded(page);

    const play4x = page.getByRole('button', { name: 'プレイ 4×' });
    await expect(play4x).toBeVisible({ timeout: 10_000 });
    await play4x.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '次のゲーム' })).toBeVisible({ timeout: 10_000 });
  });

  test('fold flow: bet → check → check → fold → result', async ({ page }) => {
    await navigateTo(page, '/ultimatetexasholdem');

    await page.getByRole('button', { name: 'ベット' }).click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'チェック' }).click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'チェック' }).click();
    await waitForLoaded(page);

    const foldButton = page.getByRole('button', { name: 'フォールド' });
    await expect(foldButton).toBeVisible({ timeout: 10_000 });
    await foldButton.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '次のゲーム' })).toBeVisible({ timeout: 10_000 });
  });
});
