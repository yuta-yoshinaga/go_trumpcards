import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Bura E2E', () => {
  test('leads a card, resolves the trick, and can be reset', async ({ page }) => {
    await navigateTo(page, '/bura');

    // The header carries the trick counter and the stock; the target is on the
    // score row. No assertion on card values -- the deal is shuffled.
    await expect(page.getByText(/トリック/).first()).toBeVisible();
    await expect(page.getByText(/目標/)).toBeVisible();

    const play = page.getByRole('button', { name: '出す' });
    await expect(play).toBeVisible();
    // Nothing is selected on arrival, so the play button starts unusable.
    await expect(play).toBeDisabled();

    // Scope to the hand container: `button[aria-pressed]` alone also matches
    // the CLI and hint toggles, and `.first()` then clicks a toggle instead of
    // a card.
    const handCards = page.locator('[data-tutorial="bura-hand"] button[aria-pressed]');
    await expect(handCards.first()).toBeVisible();
    await handCards.first().click();
    await expect(play).toBeEnabled();

    await play.click();
    await waitForLoaded(page);

    // The CPU answers and the trick resolves inside the same request, so the
    // counter has to have advanced. Asserting on the button label instead
    // would be wrong: whoever won now leads, and if that is the CPU the
    // control relabels to "N 枚で受ける".
    await expect(page.getByText(/トリック: [1-9]/)).toBeVisible();
    // Whatever the turn is, the selection was cleared by the play.
    await expect(handCards.first()).toBeVisible();
    await expect(handCards.first()).toHaveAttribute('aria-pressed', 'false');
  });

  test('claiming short ends the round immediately', async ({ page }) => {
    await navigateTo(page, '/bura');

    // A fresh deal cannot be at 31 points, so this claim is always wrong --
    // which is the point: a wrong claim has to cost the round rather than be
    // silently ignored.
    await page.getByRole('button', { name: /31点を宣言/ }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/あなたの負けです/)).toBeVisible();

    // GameResetButton relabels once the round has ended and skips the confirm
    // dialog.
    await page.getByRole('button', { name: '次のゲーム' }).click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: '出す' })).toBeVisible();
  });
});
