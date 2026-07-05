import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Panguingue (Pan) E2E', () => {
  test('loads, renders the human hand and a control, and toggles card selection', async ({ page }) => {
    await navigateTo(page, '/pan');

    // Fresh game via the mid-game reset + confirm dialog.
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Round header and score table render.
    await expect(page.getByText(/^ラウンド \d+\/\d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    // On the human's turn at least one action control is visible.
    const drawStockButton = page.getByRole('button', { name: '山札から引く' });
    const meldButton = page.getByRole('button', { name: 'メルド' });
    const discardButton = page.getByRole('button', { name: '捨てる' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });
    await expect(
      drawStockButton.or(meldButton).or(discardButton).or(nextRoundButton).or(anyResetButton).first(),
    ).toBeVisible({ timeout: 10_000 });

    // The human hand is rendered as selectable, aria-pressed toggling buttons.
    const handCards = page.locator('button[aria-pressed]:has(img)');
    await expect(handCards.first()).toBeVisible({ timeout: 10_000 });
    const firstCard = handCards.first();
    await expect(firstCard).toHaveAttribute('aria-pressed', 'false');
    await firstCard.click();
    await expect(firstCard).toHaveAttribute('aria-pressed', 'true');

    // A control stays visible after interacting.
    await expect(anyResetButton.first()).toBeVisible();
  });
});
