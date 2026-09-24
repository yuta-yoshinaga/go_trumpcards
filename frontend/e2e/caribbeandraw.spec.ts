import { expect, test } from '@playwright/test';
import { gameButton, navigateTo, waitForLoaded } from './helpers';

test.describe('Caribbean Draw Poker E2E', () => {
  test('stands pat: bet → no exchange → call → result → reset', async ({ page }) => {
    await navigateTo(page, '/caribbeandraw');

    const betButton = gameButton(page, 'ベット');
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // DRAW phase: the fee must be on screen BEFORE the player commits — it is
    // charged on confirm, and this is the only place they can see it first.
    await expect(page.getByTestId('cd-draw-fee')).toBeVisible({ timeout: 10_000 });
    const standPat = page.getByRole('button', { name: '交換しない' });
    await expect(standPat).toBeVisible();
    await standPat.click();
    await waitForLoaded(page);

    // ACTION phase
    const callButton = gameButton(page, 'コール');
    await expect(callButton).toBeVisible({ timeout: 10_000 });
    await callButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    await resetButton.click();
    await waitForLoaded(page);
    await expect(gameButton(page, 'ベット')).toBeVisible();
  });

  test('exchanges two cards, and refuses a third', async ({ page }) => {
    await navigateTo(page, '/caribbeandraw');

    await gameButton(page, 'ベット').click();
    await waitForLoaded(page);
    await expect(page.getByTestId('cd-draw-fee')).toBeVisible({ timeout: 10_000 });

    // The hand cards become toggles during the draw phase.
    const selectable = gameButton(page, /^(?!ベット|交換|コール|フォールド|リセット)/).filter({
      has: page.locator('img'),
    });
    const first = selectable.nth(0);
    const second = selectable.nth(1);
    const third = selectable.nth(2);

    await first.click();
    await second.click();
    await expect(first).toHaveAttribute('aria-pressed', 'true');
    await expect(second).toHaveAttribute('aria-pressed', 'true');

    // **The cap is two.** A third selection must not take.
    await third.click();
    await expect(third).toHaveAttribute('aria-pressed', 'false');

    await page.getByRole('button', { name: /^交換する/ }).click();
    await waitForLoaded(page);

    await expect(gameButton(page, 'コール')).toBeVisible({ timeout: 10_000 });
  });

  test('fold flow: bet → stand pat → fold → result', async ({ page }) => {
    await navigateTo(page, '/caribbeandraw');

    await gameButton(page, 'ベット').click();
    await waitForLoaded(page);
    await page.getByRole('button', { name: '交換しない' }).click();
    await waitForLoaded(page);

    const foldButton = page.getByRole('button', { name: 'フォールド' });
    await expect(foldButton).toBeVisible({ timeout: 10_000 });
    await foldButton.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '次のゲーム' })).toBeVisible({ timeout: 10_000 });
  });
});
