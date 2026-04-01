import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Euchre E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/euchre');

    // Click リセット to start
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible();
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    // Verify team scores are shown
    await expect(page.getByText('チームスコア').first()).toBeVisible();

    const orderUpButton = page.getByRole('button', { name: 'オーダーアップ', exact: true });
    const passButton = page.getByRole('button', { name: 'パス' });
    const playButton = page.getByRole('button', { name: '出す' });
    const discardButton = page.getByRole('button', { name: 'ディスカード' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 60;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        orderUpButton
          .or(passButton)
          .or(playButton)
          .or(discardButton)
          .or(nextTrickButton)
          .or(nextRoundButton)
          .or(resetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      const orderUpVisible = await orderUpButton.isVisible();
      const passVisible = await passButton.isVisible();
      const playVisible = await playButton.isVisible();
      const discardVisible = await discardButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Game end: no action buttons visible
      if (!orderUpVisible && !passVisible && !playVisible && !discardVisible && !nextTrickVisible && !nextRoundVisible)
        break;

      // Pick-up phase: order up or pass
      if (orderUpVisible) {
        interactions++;
        await orderUpButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Pass (call trump phase)
      if (passVisible && !playVisible && !discardVisible) {
        interactions++;
        // Try to select a suit button if visible, otherwise pass
        const suitButton = page.getByRole('button', { name: /♠|♣|♥|♦/ }).first();
        const suitVisible = await suitButton.isVisible();
        if (suitVisible) {
          await suitButton.click();
        } else {
          await passButton.click();
        }
        await waitForLoaded(page);
        continue;
      }

      // Discard phase
      if (discardVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await discardButton.isVisible()) && (await discardButton.isEnabled())) {
          await discardButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Play phase: select a card and play
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

      // Trick end
      if (nextTrickVisible) {
        interactions++;
        await nextTrickButton.click();
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

  test('settings: change CPU difficulty and point limit', async ({ page }) => {
    await navigateTo(page, '/euchre');

    // Open settings
    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    // Change CPU difficulty
    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // Change point limit
    const limitSelect = page.locator('select').nth(1);
    await expect(limitSelect).toBeVisible();
    await limitSelect.selectOption('21');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
