import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * The human's playable hand cards.
 *
 * **`data-legal` is what marks a playable card**, and the hand section is
 * found by `data-tutorial` — a locator on the wrong attribute matches nothing
 * and the test then "passes" having clicked nothing.
 */
function handCards(page: Parameters<typeof navigateTo>[0]) {
  return page.locator('[data-tutorial="costlycolours-player-hand"] button[data-legal]');
}

test.describe('Costly Colours E2E', () => {
  test('deals three each and turns one up', async ({ page }) => {
    await navigateTo(page, '/costlycolours');
    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。
    const turnUp = page.getByTestId('costlycolours-turnup');
    await expect(turnUp).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(turnUp.getByRole('img')).toHaveCount(1);

    // 数え上げはまだ 0。
    await expect(page.getByTestId('costlycolours-total')).toContainText('0');
  });

  // **応じる／断るは別のボタン。** 断ると相手に 1 点入るので片方を既定にしない。
  test('offers both sides of the exchange before any card is played', async ({ page }) => {
    await navigateTo(page, '/costlycolours');
    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await expect(page.getByTestId('costlycolours-mog')).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(page.getByTestId('costlycolours-nomog')).toBeVisible();
    // 交換フェーズでは札を選べない。
    await expect(handCards(page)).toHaveCount(0);
  });

  test('refusing the exchange starts the count', async ({ page }) => {
    await navigateTo(page, '/costlycolours');
    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByTestId('costlycolours-nomog').click();
    await waitForLoaded(page);

    // 数え上げに入ったら交換ボタンは消え、出せる札が並ぶ。
    await expect(page.getByTestId('costlycolours-mog')).toHaveCount(0);
    await expect(
      handCards(page)
        .or(page.getByTestId('costlycolours-next-deal'))
        .or(page.getByTestId('costlycolours-winner'))
        .first(),
    ).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('plays a card once the count has started', async ({ page }) => {
    await navigateTo(page, '/costlycolours');
    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page.getByTestId('costlycolours-nomog').click();
    await waitForLoaded(page);

    const hand = handCards(page);
    if ((await hand.count()) === 0) {
      // 珍しいが、31 を超えずに出せる札が無い配りもありうる。
      await expect(page.getByTestId('costlycolours-total')).toBeVisible({ timeout: TIMEOUT_ACTION });
      return;
    }
    await hand.first().click();
    await waitForLoaded(page);
    await expect(page.getByTestId('costlycolours-total')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/costlycolours');
    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ディール 1（61 点勝負）')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
