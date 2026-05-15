import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Oasis Poker E2E', () => {
  test('plays a round: bet → stand → call → result → reset', async ({ page }) => {
    await navigateTo(page, '/oasispoker');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // EXCHANGE phase: click ステイ (skip exchange)
    const standButton = page.getByRole('button', { name: 'ステイ' });
    await expect(standButton).toBeVisible({ timeout: 10_000 });
    await standButton.click();
    await waitForLoaded(page);

    // ACTION phase: click コール
    const callButton = page.getByRole('button', { name: 'コール' });
    await expect(callButton).toBeVisible({ timeout: 10_000 });
    await callButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });

  test('fold flow: bet → stand → fold → result → reset', async ({ page }) => {
    await navigateTo(page, '/oasispoker');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    const standButton = page.getByRole('button', { name: 'ステイ' });
    await expect(standButton).toBeVisible({ timeout: 10_000 });
    await standButton.click();
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

  test('exchange flow: bet → select card → exchange → call', async ({ page }) => {
    await navigateTo(page, '/oasispoker');

    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // EXCHANGE phase: select card 0 and click 交換
    const exchangeButton = page.getByRole('button', { name: '交換' });
    await expect(exchangeButton).toBeVisible({ timeout: 10_000 });
    await expect(exchangeButton).toBeDisabled();

    await page.getByTestId('player-card-0').click();
    await expect(exchangeButton).toBeEnabled();
    await exchangeButton.click();
    await waitForLoaded(page);

    // ACTION phase reached
    const callButton = page.getByRole('button', { name: 'コール' });
    await expect(callButton).toBeVisible({ timeout: 10_000 });
  });
});
