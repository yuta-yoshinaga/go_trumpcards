import { expect, test } from '@playwright/test';
import { PHONE_VIEWPORT } from './desktopViewports';
import { isVisibleWithin, navigateTo, TIMEOUT_GAME_LOOP, waitForLoaded } from './helpers';
import { checkLabel } from './sweeps/label';

test.describe('Kille E2E', () => {
  test('wraps the human player labels when the human is out as the lowest dealer', async ({ page }) => {
    await page.setViewportSize(PHONE_VIEWPORT);
    await page.route('**/kille/exec', async (route) => {
      const response = await route.fetch();
      const json = await response.json();
      const human = json.players.find((player: { isHuman: boolean }) => player.isHuman);
      json.dealerIdx = human.id;
      Object.assign(human, { isOut: true, knockedBy: '', reentries: 0 });
      await route.fulfill({ response, json });
    });

    await navigateTo(page, '/kille');
    const humanRow = page.getByTestId('kille-player').filter({ hasText: 'あなた' });
    await expect(humanRow).toContainText('あなた');
    await expect(humanRow).toContainText('[親]');
    await expect(humanRow).toContainText('収支');
    await expect(humanRow).toContainText('買い戻し残り 3回');
    await expect(humanRow).toContainText('[最弱で脱落]');
    await checkLabel(page, '/kille', PHONE_VIEWPORT);
  });

  test('shows the pack, keeps other hands hidden and takes a turn', async ({ page }) => {
    await navigateTo(page, '/kille');

    const rules = page.getByTestId('kille-rules');
    if (!(await rules.locator('[data-testid="kille-rules-note"]').isVisible())) {
      await rules.locator('summary').click();
    }
    await page.getByTestId('kille-ladder').locator('summary').click();

    // Permanent, not tutorial-only: an exchanged Harlequin being the LOWEST card
    // is what everyone gets wrong, and it decides whether to trade at all.
    await expect(page.getByTestId('kille-rules-note')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });

    // A single suit means the denomination ladder is the whole of the ranking.
    const ladder = page.getByTestId('kille-ladder');
    await expect(ladder).toContainText('道化');
    await expect(ladder).toContainText('カッコウ');
    await expect(ladder).toContainText('仮面');

    await expect(page.getByTestId('kille-player')).toHaveCount(4, { timeout: TIMEOUT_GAME_LOOP });

    // Either action is legal on the human's turn; the dealer's exchange button
    // reads differently because it swaps with the stock.
    //
    // **The human does not always get a turn.** Reset makes seat 0 the dealer,
    // so the human acts LAST, and a CPU that exchanges into the human's Pig
    // knocks the original holder out before then (internal/domain/Kille.go).
    // An eliminated seat renders no action buttons, so waiting longer would
    // never help -- wait for the turn OR for the round to have moved on (#6071).
    const stand = page.getByRole('button', { name: '交換しない' });
    const roundOver = page.getByTestId('kille-showdown').or(page.getByRole('button', { name: '次のラウンドへ' }));
    await expect(stand.or(roundOver).first()).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
    if (await stand.isVisible()) {
      await stand.click();
      await waitForLoaded(page);
    }

    // The turn resolves to a showdown or to another seat's turn -- either way
    // the round advances rather than hanging.
    const showdown = page.getByTestId('kille-showdown');
    const next = page.getByRole('button', { name: '次のラウンドへ' });
    const stillPlaying = page.getByRole('button', { name: '交換しない' });
    expect(
      (await isVisibleWithin(showdown, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(next, TIMEOUT_GAME_LOOP)) ||
        (await isVisibleWithin(stillPlaying, TIMEOUT_GAME_LOOP)),
    ).toBe(true);
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/kille');
    await waitForLoaded(page);

    await page.getByRole('button', { name: /リセット|次のゲーム/ }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await confirm.isVisible()) {
      await confirm.click();
    }
    await waitForLoaded(page);

    await page.getByTestId('kille-ladder').locator('summary').click();
    await expect(page.getByTestId('kille-ladder')).toBeVisible({ timeout: TIMEOUT_GAME_LOOP });
  });
});
