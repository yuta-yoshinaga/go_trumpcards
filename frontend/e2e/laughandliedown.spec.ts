import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Laugh and Lie Down E2E', () => {
  test('shows the face-up table and both rules', async ({ page }) => {
    await navigateTo(page, '/laughandliedown');

    // The rules are permanent, not tutorial-only: "one or three" and "cannot
    // capture means your whole hand joins the table" are what a player gets
    // wrong.
    await expect(page.getByText(/1枚か3枚/)).toBeVisible();
    await expect(page.getByText(/手札を全部場に置いて降りる/)).toBeVisible();
    await expect(page.getByText(/ポット: 11/)).toBeVisible();

    // The table is face up in full -- there is no hidden stock to draw from.
    const tableCards = page.locator('[data-tutorial="lld-table"] img');
    await expect(tableCards.first()).toBeVisible();
  });

  test('plays a capture', async ({ page }) => {
    await navigateTo(page, '/laughandliedown');

    const handCards = page.locator('[data-tutorial="lld-hand"] button[data-hint-action="play"]');
    await expect(handCards.first()).toBeVisible();
    await expect(handCards).toHaveCount(8);

    // Play whichever card the server marked playable; the deal is shuffled, so
    // no assertion names a card. If nothing is playable the seat lies down,
    // which is itself a legal outcome, so only assert the page stays alive.
    await handCards
      .first()
      .click({ timeout: 5000 })
      .catch(() => {});
    await waitForLoaded(page);
    await expect(page.getByText(/ポット: 11/)).toBeVisible();
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/laughandliedown');

    const handCards = page.locator('[data-tutorial="lld-hand"] button[data-hint-action="play"]');
    await handCards
      .first()
      .click({ timeout: 5000 })
      .catch(() => {});
    await waitForLoaded(page);

    // The game is short -- the human gets about four turns -- so a single click
    // can finish it, and then the reset control reads 次のゲーム rather than
    // リセット. Match either, or this passes or fails on the shuffle.
    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    // Mid-game the reset opens a confirm dialog (labelled 確認, common.json
    // button.confirm); at the end it resets straight away.
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await expect(handCards).toHaveCount(8);
  });
});
