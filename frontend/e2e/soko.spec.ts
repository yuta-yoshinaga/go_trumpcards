import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Soko E2E', () => {
  test('plays a full round: reset → check/call through streets → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/soko');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Play through the streets. Racing the three controls rather than probing
    // each in turn: sequential probes cost seconds per lap even after the hand
    // ends, which can push the loop past the test timeout and surface as a
    // bogus "browser has been closed" (#2443).
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });
      const anyControl = endResetButton.or(checkButton).or(callButton).first();
      await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

      if (await endResetButton.isVisible()) {
        roundEnded = true;
        break;
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
      await waitForLoaded(page);
    }

    expect(roundEnded).toBe(true);

    // Start next round (end state: no confirm dialog)
    await endResetButton.click();
    await waitForLoaded(page);
  });

  // The page is shared with Five Card Stud, so the thing worth pinning here is
  // that /soko really renders Soko rather than silently driving the other game.
  test('renders as Soko, not as Five Card Stud', async ({ page }) => {
    await navigateTo(page, '/soko');
    await expect(page.getByText('ソッコ').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText('ファイブカードスタッド')).toHaveCount(0);
    // Shared controls are present -- proof the reuse actually mounted.
    await expect(page.getByTestId('five-card-stud-kbd-shortcuts')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
