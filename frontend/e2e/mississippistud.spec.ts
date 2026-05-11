import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Mississippi Stud E2E', () => {
  test('plays a round: ante → 3x × 3 streets → result → reset', async ({ page }) => {
    await navigateTo(page, '/mississippistud');

    const anteButton = page.getByRole('button', { name: 'アンティ' });
    await expect(anteButton).toBeVisible();
    await anteButton.click();
    await waitForLoaded(page);

    for (let i = 0; i < 3; i++) {
      const playButton = page.getByRole('button', { name: '3倍' });
      await expect(playButton).toBeVisible({ timeout: 10_000 });
      await playButton.click();
      await waitForLoaded(page);
    }

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'アンティ' })).toBeVisible();
  });

  test('fold flow: ante → fold → result', async ({ page }) => {
    await navigateTo(page, '/mississippistud');

    const anteButton = page.getByRole('button', { name: 'アンティ' });
    await expect(anteButton).toBeVisible();
    await anteButton.click();
    await waitForLoaded(page);

    const foldButton = page.getByRole('button', { name: 'フォールド' });
    await expect(foldButton).toBeVisible({ timeout: 10_000 });
    await foldButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });
  });
});
