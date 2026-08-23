import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Ristikontra E2E', () => {
  test('deals a 2-vs-2 table and renders the pile', async ({ page }) => {
    await navigateTo(page, '/ristikontra');
    await waitForLoaded(page);

    await expect(page.getByRole('heading', { name: 'リスティコントラ' })).toBeVisible({
      timeout: 10_000,
    });

    // **Always four seats.** The clone source (Pişti) can be 2-4 players and
    // exposed a player-count selector; Ristikontra is a fixed partnership, so a
    // table of any other size would be a regression, not a setting.
    for (const seat of [0, 1, 2, 3]) {
      await expect(page.getByTestId(`ristikontra-provisional-${seat}`)).toBeVisible({
        timeout: 10_000,
      });
    }
  });

  test('plays a card and the hand keeps moving', async ({ page }) => {
    await navigateTo(page, '/ristikontra');
    await waitForLoaded(page);

    // **Assert the hand rendered before asserting anything about a play.**
    // Wrapping the play in `if (visible)` would let this pass on any run that
    // never reached the player's turn.
    const firstCard = page.getByTestId('hand-card-0');
    await expect(firstCard, 'the human hand must render before a card can be played').toBeVisible({
      timeout: 10_000,
    });
    await expect(firstCard).toBeEnabled({ timeout: 10_000 });

    await firstCard.click();
    await waitForLoaded(page);

    // The play landed and the page is still live — not stuck on an error.
    await expect(page.getByTestId('ristikontra-provisional-0')).toBeVisible({ timeout: 10_000 });
  });

  test('the table size is fixed, with no player-count selector', async ({ page }) => {
    await navigateTo(page, '/ristikontra');
    await waitForLoaded(page);

    // Pişti's settings panel offers 2/3/4 players. Carrying that control over
    // would let the player ask for a table that cannot form two teams.
    //
    // **Open the panel first.** It is a collapsed <details>, so asserting the
    // player-count control is absent while it is shut would pass for the wrong
    // reason — everything inside is absent then.
    const settings = page.locator('details').filter({ hasText: '設定' }).first();
    await settings.locator('summary').click();

    await expect(
      page.getByLabel('CPU難易度'),
      'the settings panel must actually be open, or the check below proves nothing',
    ).toBeVisible({ timeout: 10_000 });
    await expect(page.getByLabel(/プレイヤー数|人数/)).toHaveCount(0);
  });
});
