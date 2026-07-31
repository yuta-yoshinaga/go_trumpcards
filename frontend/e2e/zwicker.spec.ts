import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Zwicker E2E', () => {
  test('shows both rules and the matching values, then trails a card', async ({ page }) => {
    await navigateTo(page, '/zwicker');

    // Permanent, not tutorial-only: the two-valued courts and what a Zwick
    // actually is are the two rules a player gets wrong.
    await expect(page.getByText(/A=1\/11/)).toBeVisible();
    await expect(page.getByText(/場を空にするとZwickで1点/)).toBeVisible();

    // The human leads, so the play controls must be there.
    const trail = page.getByRole('button', { name: '置く' });
    await expect(trail).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Nothing is selected yet, so trailing is refused.
    await expect(trail).toBeDisabled();

    await page.locator('[data-hint-action="discard"]').first().click();
    await expect(trail).toBeEnabled();
    await trail.click();
    await waitForLoaded(page);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/zwicker');
    await waitForLoaded(page);

    // Several deals run before either team reaches the target, so the reset
    // control may read either label.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByText(/61点で勝ち/)).toBeVisible();
  });
});
