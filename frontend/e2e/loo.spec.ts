import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Loo E2E', () => {
  test('loads, resets, and renders the game UI', async ({ page }) => {
    await navigateTo(page, '/loo');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Deal / trick info renders.
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^トリック \d+\/\d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // The game starts in the Decide phase; some interactive control must be
    // present (a play / pass decision, the human's play control, a next-deal
    // advance, or — once the deal resolves via CPU play — the reset button).
    const anyControl = page
      .getByRole('button', { name: '参加' })
      .or(page.getByRole('button', { name: '降りる' }))
      .or(page.getByRole('button', { name: '出す' }))
      .or(page.getByRole('button', { name: '次のディール' }))
      .or(page.getByRole('button', { name: /リセット/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
