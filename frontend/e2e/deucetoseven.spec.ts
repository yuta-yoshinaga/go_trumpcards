import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('2-7 Triple Draw E2E', () => {
  test('plays a full hand: reset → check/call + stand → showdown → reset', async ({ page }) => {
    await navigateTo(page, '/deucetoseven');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    const checkButton = page.getByRole('button', { name: 'チェック', exact: true });
    const callButton = page.getByRole('button', { name: 'コール', exact: true });
    const foldButton = page.getByRole('button', { name: 'フォールド', exact: true });
    const standButton = page.getByRole('button', { name: /スタンド/ });

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

      await page.waitForTimeout(300);
    }

    expect(roundEnded).toBe(true);

    await endResetButton.click();
    await waitForLoaded(page);
  });
});
