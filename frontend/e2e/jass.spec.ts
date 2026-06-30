import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Jass (Schieber) E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/jass');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();
    await expect(page.getByText('チームスコア').first()).toBeVisible();

    const schiebenButton = page.getByRole('button', { name: 'シーバー' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const suitButton = page.getByRole('button', { name: /♠|♣|♥|♦/ }).first();
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 60;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        schiebenButton.or(suitButton).or(playButton).or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const suitVisible = await suitButton.isVisible();
      const schiebenVisible = await schiebenButton.isVisible();
      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Bid phase: pick first available suit (suit buttons appear without 出す).
      if (suitVisible && !playVisible) {
        interactions++;
        await suitButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (schiebenVisible && !playVisible && !suitVisible) {
        interactions++;
        await schiebenButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (playVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible()) && (await playButton.isEnabled())) {
          await playButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      if (nextTrickVisible) {
        interactions++;
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (nextRoundVisible) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
        continue;
      }

      break;
    }

    expect(interactions).toBeGreaterThan(0);

    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty and target score', async ({ page }) => {
    await navigateTo(page, '/jass');

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    const targetSelect = page.locator('select').nth(1);
    await expect(targetSelect).toBeVisible();
    await targetSelect.selectOption('500');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
