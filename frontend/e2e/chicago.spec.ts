import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Chicago E2E', () => {
  test('plays a full round: reset → check/call through streets → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/chicago');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through streets: THIRD_STREET → FOURTH → FIFTH → SIXTH → SEVENTH → SHOWDOWN
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });

      // Wait for whichever control appears FIRST rather than probing each one
      // in turn: two sequential 3s probes cost 6s per lap even once the hand
      // has ended, so a re-raising CPU could push 20 laps past the 90s test
      // timeout — Playwright then tears the context down and the in-flight
      // waitForSelector reports "Target page, context or browser has been
      // closed", which reads like a browser crash but is really this timeout
      // (#2443). Racing them returns as soon as any control is actionable.
      const anyControl = endResetButton.or(checkButton).or(callButton).first();
      await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

      // End state reached — break immediately.
      if (await endResetButton.isVisible()) {
        roundEnded = true;
        break;
      }

      // Try check first, then call
      if (await checkButton.isVisible()) {
        if (await checkButton.isEnabled()) {
          await checkButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      if (await callButton.isVisible()) {
        if (await callButton.isEnabled()) {
          await callButton.click();
          await waitForLoaded(page);
          continue;
        }
      }

      await waitForLoaded(page);
    }

    expect(roundEnded).toBe(true);

    // The showdown must explain the split. Either the pot halved (somebody
    // held a spade face-down) or the high took it all -- one of the two lines
    // is always there, and neither depends on the shuffle.
    const split = page.getByTestId('studchicago-split');
    const hiTakesAll = page.getByTestId('studchicago-hi-takes-all');
    await expect(split.or(hiTakesAll).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // Start next round (end state: no confirm dialog)
    await endResetButton.click();
    await waitForLoaded(page);
  });
});
