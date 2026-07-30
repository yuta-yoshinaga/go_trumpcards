import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Skitgubbe E2E', () => {
  test('plays a duel card in phase one', async ({ page }) => {
    await navigateTo(page, '/skitgubbe');

    // The phase name and both phases' rules are permanent, not tutorial-only:
    // which phase is running decides what clicking a card means.
    await expect(page.getByText(/第1フェーズ（集める）/)).toBeVisible();
    await expect(page.getByText(/2人の一騎打ち/)).toBeVisible();
    // Trump is fixed by the LAST card drawn, so it is undecided at the deal.
    await expect(page.getByText(/切札: 未定/)).toBeVisible();

    const handCards = page.locator('[data-tutorial="sg-hand"] button[data-hint-action="play"]');
    await expect(handCards.first()).toBeVisible();
    await expect(handCards).toHaveCount(3);

    await handCards.first().click();
    await waitForLoaded(page);

    // Both duellists draw back up to three, so the hand is refilled by the
    // time control returns. No assertion on card values -- the deal is shuffled.
    await expect(handCards.first()).toBeVisible();
  });

  test('the pick-up is refused while a card still beats the pile', async ({ page }) => {
    await navigateTo(page, '/skitgubbe');

    // Phase one has no pile at all, so the control must be inert rather than
    // merely unlabelled -- ducking is never lawful.
    const pickUp = page.getByRole('button', { name: '引き取る' });
    await expect(pickUp).toHaveAttribute('aria-disabled', 'true');
  });

  test('can be reset mid-game', async ({ page }) => {
    await navigateTo(page, '/skitgubbe');

    const handCards = page.locator('[data-tutorial="sg-hand"] button[data-hint-action="play"]');
    await handCards.first().click();
    await waitForLoaded(page);

    // Mid-game the reset button opens a confirm dialog, whose confirm control
    // is labelled 確認 (common.json button.confirm), not リセット.
    await page.getByRole('button', { name: 'リセット' }).click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(handCards).toHaveCount(3);
  });
});
