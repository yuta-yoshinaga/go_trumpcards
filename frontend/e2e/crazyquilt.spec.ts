import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('CrazyQuilt E2E', () => {
  test('navigates, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/crazyquilt');

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

  test('renders a 64-cell quilt and marks only the takeable cards', async ({ page }) => {
    await navigateTo(page, '/crazyquilt');

    // Scoped to the heading role: the NavBar lists every game, so an unscoped
    // text match finds other games' names on a correct page.
    await expect(page.getByRole('heading', { name: 'クレイジーキルト' })).toBeVisible();

    await expect(page.getByTestId('cq-cell-63')).toBeVisible();
    await expect(page.getByTestId('cq-cell-64')).toHaveCount(0);

    // An intact quilt exposes exactly sixteen short sides -- the rule the whole
    // game turns on, and the one #5274 got wrong.
    await expect(page.locator('[data-available="true"]')).toHaveCount(16);
  });

  // #5288 shipped a game whose Go zone struct and client disagreed, so every
  // move with a destination came back "param error: ..." and only E2E saw it.
  // A rule rejection is a normal Japanese message; a plumbing failure is that
  // English string, so asserting its absence catches the class without
  // depending on which cards were dealt.
  test('a move with a destination reaches the domain, not a param error', async ({ page }) => {
    await navigateTo(page, '/crazyquilt');

    await page.locator('[data-available="true"]').first().click();
    await waitForLoaded(page);

    await expect(page.getByText(/param error/i)).toHaveCount(0);
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('drawing moves a card out of the stock', async ({ page }) => {
    await navigateTo(page, '/crazyquilt');

    // The deal is shuffled, so assert the counter moves rather than which card.
    const stock = page.getByRole('button', { name: /山札 残り/ });
    const before = await stock.textContent();
    await stock.click();
    await waitForLoaded(page);
    await expect(stock).not.toHaveText(before ?? '');
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/crazyquilt');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント', exact: true })).not.toBeVisible();
  });
});
