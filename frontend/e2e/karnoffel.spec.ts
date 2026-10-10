import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';

test.describe('Karnöffel E2E', () => {
  test('shows the irregular ranking and progresses the hand', async ({ page }) => {
    await navigateTo(page, '/karnoffel');

    // Desktop details start open; only open them when running at a closed viewport.
    for (const testId of ['karnoffel-ladder-summary', 'karnoffel-chosen-note-summary']) {
      const summary = page.getByTestId(testId);
      if (!(await summary.locator('xpath=..').evaluate((details) => details.hasAttribute('open')))) {
        await summary.click();
      }
    }
    const ladder = page.getByTestId('karnoffel-ladder');
    await expect(ladder).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    await expect(ladder).toContainText('J（カルニッフェル）');
    await expect(ladder).toContainText('7（悪魔・リード時のみ）');
    await expect(page.getByTestId('karnoffel-chosen-note')).toContainText('最も低い札');

    await expect(page.getByTestId('karnoffel-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });
    await expect(page.getByTestId('karnoffel-scores')).toBeVisible();

    // The hand must progress rather than hang.
    const play = page.getByRole('button', { name: '出す' });
    const next = page.getByRole('button', { name: '次の局へ' });
    expect((await isVisibleWithin(play, TIMEOUT_GAME_LOOP)) || (await isVisibleWithin(next, TIMEOUT_GAME_LOOP))).toBe(
      true,
    );
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/karnoffel');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    const summary = page.getByTestId('karnoffel-ladder-summary');
    if (!(await summary.locator('xpath=..').evaluate((details) => details.hasAttribute('open')))) {
      await summary.click();
    }
    await expect(page.getByTestId('karnoffel-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });

  test('applies the selected target hands when resetting', async ({ page }) => {
    await navigateTo(page, '/karnoffel');
    await waitForLoaded(page);

    const settingsSummary = page.getByTestId('karnoffel-settings-summary');
    if (!(await settingsSummary.locator('xpath=..').evaluate((details) => details.hasAttribute('open')))) {
      await settingsSummary.click();
    }
    await page.getByLabel('目標局数').selectOption('5');
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) await confirm.click();

    await expect(page.getByTestId('karnoffel-scores')).toContainText('5局先取', { timeout: TIMEOUT_GAME_LOOP });
  });
});
