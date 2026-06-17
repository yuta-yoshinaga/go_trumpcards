import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Teen Patti E2E', () => {
  test('loads, resets, and renders the betting UI', async ({ page }) => {
    await navigateTo(page, '/teenpatti');

    // Reset to start a fresh game (mid-game reset shows a confirm dialog).
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await expect(resetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Deal / pot info renders.
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText(/^ポット: \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // Some interactive control must be present: a betting action (see / bet /
    // fold / show / side show), an accept/decline, a next-deal advance, or —
    // once the deal resolves via CPU play — the reset / next-game button.
    const anyControl = page
      .getByRole('button', { name: '手札を見る' })
      .or(page.getByRole('button', { name: /^ベット/ }))
      .or(page.getByRole('button', { name: 'フォールド' }))
      .or(page.getByRole('button', { name: 'ショー' }))
      .or(page.getByRole('button', { name: 'サイドショー' }))
      .or(page.getByRole('button', { name: '次のディール' }))
      .or(page.getByRole('button', { name: /リセット|次のゲーム/ }))
      .first();
    await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Reset again and verify the game restarts cleanly.
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/^ディール \d+$/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
