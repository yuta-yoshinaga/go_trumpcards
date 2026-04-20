import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Badugi E2E', () => {
  test('plays a full hand: reset → check/call + stand → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/badugi');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through up to 4 betting rounds + 3 draw rounds.
    // Each iteration waits (once) for any actionable button to appear, then acts.
    // Using a single .or() locator avoids burning time with per-button timeouts
    // that compound past the 90s test budget in CI.
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
    const callButton = page.getByRole('button', { name: 'コール', exact: true });
    const standButton = page.getByRole('button', { name: 'スタンド', exact: true });
    const anyButton = endResetButton.or(standButton).or(checkButton).or(callButton);

    let roundEnded = false;
    for (let round = 0; round < 30; round++) {
      // Wait for any actionable button to become visible (CPU may be thinking).
      await expect(anyButton.first()).toBeVisible({ timeout: 10_000 });

      if (await endResetButton.isVisible()) {
        roundEnded = true;
        break;
      }

      // Draw phase: スタンド (no exchanges).
      if ((await standButton.isVisible()) && (await standButton.isEnabled())) {
        await standButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Betting phase: prefer check, fall back to call.
      if ((await checkButton.isVisible()) && (await checkButton.isEnabled())) {
        await checkButton.click();
        await waitForLoaded(page);
        continue;
      }
      if ((await callButton.isVisible()) && (await callButton.isEnabled())) {
        await callButton.click();
        await waitForLoaded(page);
        continue;
      }

      // None of the buttons are actionable right now — let the page settle.
      await waitForLoaded(page);
    }

    expect(roundEnded).toBe(true);

    await endResetButton.click();
    await waitForLoaded(page);
  });
});
