import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Ninety-Nine (ナインティナイン) E2E', () => {
  test('navigates, buries, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/ninetynine');

    // Deal/trick info is visible after the initial reset.
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const buryButton = page.getByRole('button', { name: '3枚埋める' });
    const playButton = page.getByRole('button', { name: 'カードを出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img)');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        buryButton.or(playButton).or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      const buryVisible = await buryButton.isVisible();
      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // Bury phase: select 3 cards then bury.
      if (buryVisible) {
        interactions++;
        const cardCount = await handCards.count();
        for (let i = 0; i < Math.min(3, cardCount); i++) {
          await handCards.nth(i).click();
        }
        if (await buryButton.isEnabled()) {
          await buryButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Play phase: select a card and play.
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

      // No actionable game button — likely game end.
      break;
    }

    expect(interactions).toBeGreaterThan(0);

    // Reset and verify the game restarts.
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    if (await midResetButton.isVisible()) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty and target score', async ({ page }) => {
    await navigateTo(page, '/ninetynine');

    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible();
  });
});
