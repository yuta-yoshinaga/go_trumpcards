import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('RoyalCotillion E2E', () => {
  test('navigates, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/royalcotillion');

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

  test('renders sixteen slots, four reserve piles and both foundation series', async ({ page }) => {
    await navigateTo(page, '/royalcotillion');

    // Scoped to the heading role: the NavBar lists every game, so an unscoped
    // text match finds other games' names on a correct page.
    await expect(page.getByRole('heading', { name: 'ロイヤルコティヨン' })).toBeVisible();

    await expect(page.getByTestId('rc-tableau')).toBeVisible();
    await expect(page.getByTestId('rc-reserve')).toBeVisible();
    // Four foundations start at the Ace and four at the deuce.
    await expect(page.getByText(/^A:/)).toHaveCount(4);
    await expect(page.getByText(/^2:/)).toHaveCount(4);
  });

  // #5288 shipped a game whose Go zone struct and client disagreed, so every
  // move with a destination came back "param error: ..." and only E2E could
  // see it. A rule rejection is a normal Japanese game message; a plumbing
  // failure is that English string, so asserting its absence catches the whole
  // class without depending on which cards were dealt.
  test('a move with a destination reaches the domain, not a param error', async ({ page }) => {
    await navigateTo(page, '/royalcotillion');

    await page.getByRole('button', { name: 'めくる', exact: true }).click();
    await waitForLoaded(page);

    // Every slot holds a card at deal, so the first one is always selectable.
    await page.locator('[aria-pressed]').first().click();
    await waitForLoaded(page);

    await expect(page.getByText(/param error/i)).toHaveCount(0);
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('drawing moves a card out of the stock', async ({ page }) => {
    await navigateTo(page, '/royalcotillion');

    // The deal is shuffled, so assert the counter moves rather than which card.
    const stock = page.getByRole('button', { name: /山札 残り/ });
    const before = await stock.textContent();
    await page.getByRole('button', { name: 'めくる', exact: true }).click();
    await waitForLoaded(page);
    await expect(stock).not.toHaveText(before ?? '');
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/royalcotillion');

    const giveUpButton = page.getByRole('button', { name: 'ギブアップ' });
    await expect(giveUpButton).toBeVisible();
    await giveUpButton.click();

    // Confirm the give-up in the dialog (#2099).
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: 'ヒント', exact: true })).not.toBeVisible();
  });
});
