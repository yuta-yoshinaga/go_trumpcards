import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Doppelkopf E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/doppelkopf');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round / trick info is visible.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const announceButton = page.getByRole('button', { name: /宣言/ });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(playButton.or(nextTrickButton).or(nextRoundButton).or(anyResetButton).first()).toBeVisible({
        timeout: 10_000,
      });

      // Play phase.
      if (await playButton.isVisible()) {
        interactions++;
        const count = await handCards.count();
        if (count > 0) await handCards.first().click();
        if (await playButton.isEnabled()) {
          await playButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      if (await nextTrickButton.isVisible()) {
        interactions++;
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      if (await nextRoundButton.isVisible()) {
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Only a reset/next-game button left — the game reached an end state.
      break;
    }

    // The optional announce button never blocks progress; touch it if present.
    void announceButton;

    expect(interactions).toBeGreaterThan(0);

    // Reset and verify the game restarts.
    if (await midResetButton.isVisible()) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
