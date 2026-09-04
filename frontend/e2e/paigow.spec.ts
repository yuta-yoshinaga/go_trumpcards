import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Pai Gow Poker E2E', () => {
  test('plays a round: bet → set hands → result → reset', async ({ page }) => {
    await navigateTo(page, '/paigow');

    // BET phase: click ベット (exact match to avoid matching the ChipBetInput ± steppers)
    const betButton = page.getByRole('button', { name: 'ベット', exact: true });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // SET HANDS phase: select 2 cards and click セット
    // Cards should be visible; click on the first two to select them for low hand.
    // Cards are labelled with their suit/rank (cardAlt), so select by the toggle-state attribute.
    const cards = page.locator('[data-tutorial="pg-set-hands"] button[aria-pressed]');
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });

    // **先頭2枚を決め打ちしない。** セットは foul (2枚の下位手が5枚の上位手を
    // 上回る) の間 disabled のままで、先頭2枚が合法な分割になるかは配り次第。
    // `toBeVisible()` は disabled でも通るので、失敗は次の click() が
    // actionability を 90 秒待って timeout する形で出ていた (#6689)。
    // ハウスウェイがある以上、合法な組み合わせは必ず1つはある。
    const setButton = page.getByTestId('set-hands-button');
    await expect(setButton).toBeVisible();

    const cardCount = await cards.count();
    let chosen: [number, number] | null = null;
    for (let i = 0; i < cardCount && chosen === null; i += 1) {
      for (let j = i + 1; j < cardCount && chosen === null; j += 1) {
        await cards.nth(i).click();
        await cards.nth(j).click();
        if (await setButton.isEnabled()) {
          chosen = [i, j];
        } else {
          // 選択はトグル。押した2枚だけ戻せば次の組に移れる。
          await cards.nth(i).click();
          await cards.nth(j).click();
        }
      }
    }
    expect(chosen, '合法な2枚の組み合わせが1つも見つからなかった').not.toBeNull();

    await setButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット', exact: true })).toBeVisible();
  });
});
