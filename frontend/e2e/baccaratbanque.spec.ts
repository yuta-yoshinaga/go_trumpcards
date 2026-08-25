import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The banker's two decision buttons.
 *
 * **They are keyed by `aria-keyshortcuts`, not by their label.** "引く" also
 * appears in the guidance line and in the tutorial copy, so a name locator
 * matches several nodes and `.first()` can land on text that is not a button.
 */
function drawButton(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('button[aria-keyshortcuts="d"]');
}
function standButton(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('button[aria-keyshortcuts="s"]');
}

test.describe('Baccarat Banque E2E', () => {
  test('deals one bank against two tableaux, all face up', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **3 席とも表向き。** バカラに伏せ札は無い。
    for (const role of ['banker', 'right', 'left']) {
      await expect(page.getByTestId(`banque-hand-${role}`)).toBeVisible({ timeout: TIMEOUT_ACTION });
      await expect(page.getByTestId(`banque-total-${role}`)).toBeVisible();
    }
    await expect(page.getByTestId('banque-shoe-line')).toBeVisible();
  });

  // **負けても席が動かないことは残高から読めない。** 盤に書いてあること。
  test('says the bank survives a loss', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toContainText('負けても続きます', {
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('offers both decisions, and exactly one button each', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 親がナチュラルだと判断そのものが無く、そのまま決着まで進む配りもある。
    if (await isVisibleWithin(page.getByTestId('banque-free-notice'), TIMEOUT_ACTION)) {
      await expect(drawButton(page)).toHaveCount(1);
      await expect(standButton(page)).toHaveCount(1);
      await expect(page.getByTestId('banque-free-notice')).toContainText('固定表はありません');
      return;
    }
    await expect(page.getByTestId('banque-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('standing settles both tableaux separately', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    if (await isVisibleWithin(standButton(page), TIMEOUT_ACTION)) {
      await standButton(page).click();
      await waitForLoaded(page);
    }

    await expect(page.getByTestId('banque-result')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // **左右は 1 行ずつ。** 差額だけにまとめない。
    await expect(page.getByTestId('banque-side-1')).toBeVisible();
    await expect(page.getByTestId('banque-side-2')).toBeVisible();
    await expect(page.getByTestId('banque-net')).toBeVisible();
  });

  test('keeps the bank across the next coup', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    if (await isVisibleWithin(standButton(page), TIMEOUT_ACTION)) {
      await standButton(page).click();
      await waitForLoaded(page);
    }
    await page.getByRole('button', { name: '次のクー' }).click();
    await waitForLoaded(page);

    // 勝っていても負けていても、次のクーはあなたのバンクのまま 2 回目になる。
    await expect(page.getByTestId('banque-coup-line')).toContainText('2', { timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/baccaratbanque');
    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('banque-coup-line')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
