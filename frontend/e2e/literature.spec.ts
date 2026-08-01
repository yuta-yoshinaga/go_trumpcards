import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Literature E2E', () => {
  test('states the real threshold and offers only opponents', async ({ page }) => {
    await navigateTo(page, '/literature');

    // Permanent, not tutorial-only: five of eight wins, and four decides nothing.
    const note = page.getByTestId('literature-threshold-note');
    await expect(note).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(note).toContainText('8組の過半数=5組');
    await expect(page.getByTestId('literature-hidden-note')).toContainText('味方の手札も見えません');

    await expect(page.getByTestId('literature-player')).toHaveCount(6, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('literature-halfsuits')).toBeVisible();

    // The human opens, so the controls are there and only opponents are offered.
    const ask = page.getByRole('button', { name: '要求する' });
    if (await isVisibleWithin(ask, TIMEOUT_GAME_LOOP)) {
      await expect(page.getByTestId('literature-ask-rules')).toContainText('相手チームにのみ');
      await ask.click();
      await waitForLoaded(page);
    }

    // The game must progress rather than hang.
    const claim = page.getByRole('button', { name: '宣言する' });
    expect((await isVisibleWithin(ask, TIMEOUT_GAME_LOOP)) || (await isVisibleWithin(claim, TIMEOUT_GAME_LOOP))).toBe(
      true,
    );
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/literature');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(page.getByTestId('literature-threshold-note')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
