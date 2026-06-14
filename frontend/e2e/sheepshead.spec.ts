import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Sheepshead E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/sheepshead');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round / trick info is visible.
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText(/^トリック \d+$/).first()).toBeVisible();

    const pickButton = page.getByRole('button', { name: 'ピックする' });
    const passButton = page.getByRole('button', { name: 'パスする' });
    const buryButton = page.getByRole('button', { name: /埋める/ });
    const callButton = page.getByRole('button', { name: /を呼ぶ$/ });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    const MAX_TURNS = 80;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect(
        pickButton
          .or(passButton)
          .or(buryButton)
          .or(callButton)
          .or(playButton)
          .or(nextTrickButton)
          .or(nextRoundButton)
          .or(anyResetButton)
          .first(),
      ).toBeVisible({ timeout: 10_000 });

      // Pick phase: take the blind to keep the game moving.
      if (await pickButton.isVisible()) {
        interactions++;
        await pickButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Bury phase: select two cards then bury.
      if (await buryButton.isVisible()) {
        interactions++;
        const count = await handCards.count();
        if (count >= 2) {
          await handCards.nth(0).click();
          await handCards.nth(1).click();
        }
        if (await buryButton.isEnabled()) {
          await buryButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Call phase: pick the first callable suit.
      if (await callButton.first().isVisible()) {
        interactions++;
        await callButton.first().click();
        await waitForLoaded(page);
        continue;
      }

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
