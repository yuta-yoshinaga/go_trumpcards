import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Trex E2E', () => {
  test('shows both rules and lets the king choose a contract', async ({ page }) => {
    await navigateTo(page, '/trex');

    // Permanent, not tutorial-only: "once per kingdom" and "the dominoes start
    // from the JACK" are what a player gets wrong.
    await expect(page.getByText(/同じ契約は1王国に1度だけ/)).toBeVisible();
    await expect(page.getByText(/Jを起点/)).toBeVisible();
    await expect(page.getByText(/ディール: 0\/20/)).toBeVisible();

    // The human is king only when dealt the seven of hearts. When a CPU is
    // king the interactor resolves its choice inside the same request, so the
    // page arrives already playing -- "waiting for the king" is a fallback the
    // web flow normally skips, and depending on it made this test fail.
    const contractButton = page.getByRole('button', { name: /♥K|ダイヤ|クイーン|トリック|ドミノ/ }).first();
    const handCard = page.locator('[data-hint-action="play"]').first();
    await expect(contractButton.or(handCard).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    if (await contractButton.isVisible()) {
      await contractButton.click();
      await waitForLoaded(page);
    }
    // Either way a contract is now in play, so the header has left "not chosen".
    await expect(page.getByText(/契約: 未選択/)).toHaveCount(0);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/trex');

    const contractButton = page.getByRole('button', { name: /♥K|ダイヤ|クイーン|トリック|ドミノ/ }).first();
    if (await contractButton.isVisible()) {
      await contractButton.click();
      await waitForLoaded(page);
    }

    // Twenty deals run long, so the reset control may read either label.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/ディール: 0\/20/)).toBeVisible();
  });
});
