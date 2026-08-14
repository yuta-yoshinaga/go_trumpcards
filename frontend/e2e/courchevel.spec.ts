import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_GAME_LOOP, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Counts what the community-card area is showing.
 *
 * **Scoped to the board.** The human's own hole cards are face up too, so a
 * page-wide count of face-up cards is satisfied by any Omaha game and proves
 * nothing about the one rule this game has.
 */
async function board(page: Parameters<typeof navigateTo>[0]) {
  const area = page.locator('[data-tutorial="bo-community-cards"]').first();
  await expect(area).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  return {
    faceUp: await area.locator('img:not([alt="カード裏面"])').count(),
    backs: await area.getByAltText('カード裏面').count(),
  };
}

/** Resets to a fresh hand and waits for the board to be up. */
async function freshHand(page: Parameters<typeof navigateTo>[0], path: string) {
  await navigateTo(page, path);
  const midResetButton = page.getByRole('button', { name: 'リセット' });
  await expect(midResetButton).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  await midResetButton.click();
  await page.getByRole('button', { name: '確認' }).click();
  await waitForLoaded(page);
  await expect(page.getByText('コミュニティカード')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
}

test.describe('Courchevel E2E', () => {
  // **プリフロップで場が 1 枚だけ見えている。** これがこのゲームの全部。
  test('shows exactly one board card before the opening bet', async ({ page }) => {
    await freshHand(page, '/courchevel');
    const { faceUp, backs } = await board(page);
    expect(faceUp).toBe(1);
    expect(backs).toBe(0);
  });

  // **同じ検査を Big O にかけると逆になる。** これが無いと、上の検査が
  // 「オマハなら何でも通る」ものでないことを示せない。
  test('big o still hides the whole board at the same point', async ({ page }) => {
    await freshHand(page, '/bigo');
    const { faceUp, backs } = await board(page);
    expect(faceUp).toBe(0);
    expect(backs).toBe(5);
  });

  test('plays a full round: reset → check/call through rounds → showdown → reset', async ({ page }) => {
    await freshHand(page, '/courchevel');

    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    let roundEnded = false;
    for (let round = 0; round < 20; round++) {
      const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
      const callButton = page.getByRole('button', { name: 'コール', exact: true });

      // どれか 1 つが操作可能になるまで待つ (順に probe すると 1 周 6 秒かかり、
      // 再レイズする CPU がいると 90 秒のテスト上限を越える。#2443)
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

    await endResetButton.click();
    await waitForLoaded(page);
  });
});
