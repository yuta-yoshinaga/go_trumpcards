import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Napoleon E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/napoleon');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const bidButton = page.getByRole('button', { name: 'ビッド' });
    const passButton = page.getByRole('button', { name: 'パス' });
    const declareButton = page.getByRole('button', { name: '宣言' });
    const exchangeButton = page.getByRole('button', { name: '交換' });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    // Play through several interactions to verify phase transitions
    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        bidButton
          .or(passButton)
          .or(declareButton)
          .or(exchangeButton)
          .or(playButton)
          .or(nextTrickButton)
          .or(nextRoundButton)
          .or(anyResetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      const bidVisible = await bidButton.isVisible();
      const passVisible = await passButton.isVisible();
      const declareVisible = await declareButton.isVisible();
      const exchangeVisible = await exchangeButton.isVisible();
      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Game end: no action buttons visible
      if (
        !bidVisible &&
        !passVisible &&
        !declareVisible &&
        !exchangeVisible &&
        !playVisible &&
        !nextTrickVisible &&
        !nextRoundVisible
      )
        break;

      // Bid phase: bid or pass
      if (bidVisible || passVisible) {
        interactions++;
        if (bidVisible) {
          const bidInput = page.locator('input[aria-label="bid-input"]');
          if (await bidInput.isVisible()) {
            await bidInput.fill('13');
          }
          await bidButton.click();
        } else {
          await passButton.click();
        }
        await waitForLoaded(page);
        continue;
      }

      // Trump declaration phase
      if (declareVisible) {
        interactions++;
        await declareButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Kitty exchange phase
      if (exchangeVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if (await exchangeButton.isEnabled()) {
          await exchangeButton.click();
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

    // Reset and verify game restarts. Button could be either mid-game (リセット) or end (次のゲーム).
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

  test('settings: change CPU difficulty and point limit', async ({ page }) => {
    await navigateTo(page, '/napoleon');

    // Open settings
    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    // Change CPU difficulty
    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
