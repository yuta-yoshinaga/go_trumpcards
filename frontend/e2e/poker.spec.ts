import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Poker E2E', () => {
  test('plays a full round: reset → bet → stand (no exchange) → bet → result → reset', async ({ page }) => {
    await navigateTo(page, '/poker');

    // Click リセット to start a new game (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
    const callButton = page.getByRole('button', { name: 'コール', exact: true });
    const standButton = page.getByRole('button', { name: 'スタンド' });
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });

    // One loop drives the whole hand — deal betting, exchange, and second
    // betting — answering whichever control the game offers next.
    //
    // It used to be a step per phase, each probing its own control with
    // `if (await isVisibleWithin(btn, TIMEOUT_ACTION))`. That is what made this
    // spec flaky (#4459). A 3s probe loses the race whenever the first betting
    // round runs long — three CPUs re-raising each other takes well over 3s —
    // and the `if` then turns a *mandatory* action into a silent no-op. The
    // hand cannot advance until the human stands or exchanges, so the run died
    // ten seconds later in the betting loop, waiting for チェック/コール/次のゲーム
    // that could never appear. The CI trace showed the page parked on
    // 「現在のフェーズ: 交換、あなたのターン」 with スタンド visible throughout.
    //
    // Racing every control the human can be asked for removes the timing
    // assumption entirely: order and duration no longer matter, only that the
    // game is waiting for one of them. A real stall still fails fast, because
    // no control appearing within TIMEOUT_GAME_LOOP is itself the failure.
    //
    // Keep the cap well above the ~8 decisions a 3-CPU hand needs. Looping
    // without one risks outliving the 90s test timeout, which surfaces as
    // "Target page, context or browser has been closed" and reads like a
    // browser crash rather than a timeout (#2443).
    for (let i = 0; i < 24; i++) {
      const anyControl = endResetButton.or(standButton).or(checkButton).or(callButton).first();
      await expect(anyControl).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
      if (await endResetButton.isVisible()) break;
      if (await standButton.isVisible()) {
        await standButton.click(); // decline the exchange, play the dealt hand
      } else if (await checkButton.isVisible()) {
        await checkButton.click();
      } else {
        await callButton.click();
      }
      await waitForLoaded(page);
    }

    // END phase: 次のゲーム should be visible
    await expect(endResetButton).toBeVisible({ timeout: 10_000 });

    // Start another round (end state: no confirm dialog)
    await endResetButton.click();
    await waitForLoaded(page);
  });
});
