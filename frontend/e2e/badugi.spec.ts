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
    const foldButton = page.getByRole('button', { name: 'フォールド', exact: true });
    const standButton = page.getByRole('button', { name: 'スタンド', exact: true });

    // Up to 4 bet rounds + 3 draw rounds × (1 human + 3 CPU) polls. Budget
    // is generous since CI scheduling can stretch CPU turns.
    let roundEnded = false;
    for (let round = 0; round < 120; round++) {
      if (await endResetButton.isVisible()) {
        roundEnded = true;
        break;
      }
      if ((await standButton.isVisible()) && (await standButton.isEnabled())) {
        await standButton.click();
        await waitForLoaded(page);
        continue;
      }
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
      if ((await foldButton.isVisible()) && (await foldButton.isEnabled())) {
        await foldButton.click();
        await waitForLoaded(page);
        continue;
      }

      // No action currently available — CPU probably thinking. Brief sleep
      // then re-check; aria-busy alone is not a sufficient signal.
      await page.waitForTimeout(300);
    }

    expect(roundEnded).toBe(true);

    await endResetButton.click();
    await waitForLoaded(page);
  });
});
