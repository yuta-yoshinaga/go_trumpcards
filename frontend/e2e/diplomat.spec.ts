import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Diplomat E2E', () => {
  test('navigates, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/diplomat');

    await expect(page.getByText(/手数/)).toBeVisible();

    const drawButton = page.getByRole('button', { name: 'めくる', exact: true });
    await expect(drawButton).toBeVisible();
    await drawButton.click();
    await waitForLoaded(page);

    const hintButton = page.getByRole('button', { name: 'ヒント', exact: true });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/手数/)).toBeVisible();
  });

  // #5288 shipped a game whose Go zone struct said `col` while the client sent
  // `idx`, so every move with a destination came back
  // "param error: to.col is required." and only E2E could see it. A rule
  // rejection is a normal Japanese game message; a plumbing failure is that
  // English string, so asserting its absence catches the whole class without
  // depending on which cards were dealt.
  test('a move with a destination reaches the domain, not a param error', async ({ page }) => {
    await navigateTo(page, '/diplomat');

    await page.getByRole('button', { name: 'めくる', exact: true }).click();
    await waitForLoaded(page);

    // Column 0 always holds four cards at deal, so its top is always selectable.
    const firstCard = page.locator('[aria-pressed]').first();
    await firstCard.click();
    await waitForLoaded(page);

    await expect(page.getByText(/param error/i)).toHaveCount(0);
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('drawing moves a card out of the stock', async ({ page }) => {
    await navigateTo(page, '/diplomat');

    // The deal is shuffled, so assert the counter moves rather than which card.
    const stock = page.getByRole('button', { name: /山札 残り/ });
    const before = await stock.textContent();
    await page.getByRole('button', { name: 'めくる', exact: true }).click();
    await waitForLoaded(page);
    await expect(stock).not.toHaveText(before ?? '');
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/diplomat');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント', exact: true })).not.toBeVisible();
  });
});
