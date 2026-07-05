import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Machiavelli E2E', () => {
  test('loads, resets, and renders the human turn controls', async ({ page }) => {
    await navigateTo(page, '/machiavelli');

    // Start (mid-game reset -> confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round header, stock and the shared table render.
    await expect(page.getByText(/^ラウンド \d+\/\d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();
    await expect(page.getByText('場のメルド', { exact: true }).first()).toBeVisible();

    const drawButton = page.getByRole('button', { name: '山札から引く' });
    const newMeldButton = page.getByRole('button', { name: 'メルドを出す' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    // On the human's turn the draw + new-meld controls are visible.
    await expect(drawButton.or(nextRoundButton).or(anyResetButton).first()).toBeVisible({ timeout: 10_000 });

    if (await drawButton.isVisible()) {
      await expect(newMeldButton).toBeVisible();

      // Selecting a hand card toggles its aria-pressed state.
      const cardCount = await handCards.count();
      if (cardCount > 0) {
        const firstCard = handCards.first();
        await expect(firstCard).toHaveAttribute('aria-pressed', 'false');
        await firstCard.click();
        await expect(firstCard).toHaveAttribute('aria-pressed', 'true');
        await firstCard.click();
        await expect(firstCard).toHaveAttribute('aria-pressed', 'false');
      }

      // Drawing a card advances the turn.
      await drawButton.click();
      await waitForLoaded(page);
    }

    // A reset control stays available so the game can always be restarted.
    await expect(anyResetButton.first()).toBeVisible();
  });
});
