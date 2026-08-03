import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Guandan E2E', () => {
  test('explains the level cards and plays a combination', async ({ page }) => {
    await navigateTo(page, '/guandan');

    // Permanent, not tutorial-only: the level cards beat aces, and the climb is 1/2/4.
    const levelNote = page.getByTestId('guandan-level-note');
    await expect(levelNote).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(levelNote).toContainText('Aより強く');
    await expect(page.getByTestId('guandan-advance-note')).toContainText('3段階上がることはありません');

    await expect(page.getByTestId('guandan-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('guandan-table')).toBeVisible();

    // Playing needs a selection first — the button stays disabled until then.
    const play = page.getByRole('button', { name: '出す' });
    if (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) {
      await expect(play).toBeDisabled();
      await page.getByTestId('hand-card-0').click();
      await expect(play).toBeEnabled();
      await play.click();
      await waitForLoaded(page);
    }

    // The game must progress rather than hang.
    const pass = page.getByRole('button', { name: 'パス' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect(
      (await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(pass, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/guandan');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('guandan-level-note')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
