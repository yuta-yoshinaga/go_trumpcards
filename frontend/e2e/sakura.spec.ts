import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Plays one hand card. When the played card matches two field cards the page
 * asks which one to take, so the highlighted field card is clicked next.
 *
 * Returns false when there was nothing to play (a CPU is thinking, or the round
 * has ended).
 */
async function playOneCard(page: Parameters<typeof navigateTo>[0]) {
  const card = page.getByTestId('hand-card-0');
  if (!(await isVisibleWithin(card, TIMEOUT_ACTION))) return false;
  if (await card.isDisabled()) return false;
  await card.click();

  // 2 枚一致のときだけ場札を選ぶ画面になる。
  const pick = page.getByTestId('sakura-field-pick');
  if (await isVisibleWithin(pick, TIMEOUT_ACTION)) {
    await page.locator('[data-capture-candidate="true"]').first().click();
  }
  await waitForLoaded(page);
  return true;
}

test.describe('Sakura E2E', () => {
  test('deals seven cards, six field cards and hides the other hands', async ({ page }) => {
    await navigateTo(page, '/sakura');
    await expect(page.getByTestId('hand-card-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **自分の手札は 7 枚、場は 6 枚。**
    await expect(page.getByTestId(/^hand-card-/)).toHaveCount(7);
    await expect(page.getByTestId(/^field-card-/)).toHaveCount(6);

    // **相手の手札は届かない。** 席の行だけが出る。
    await expect(page.getByTestId('sakura-seat-1')).toBeVisible();
    await expect(page.getByTestId('sakura-seat-2')).toBeVisible();
  });

  // **手番を打つと手札が減る。** 席の点数は取れなければ動かないので、
  // 必ず変わるのは手札の枚数のほう。
  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/sakura');
    await expect(page.getByTestId('hand-card-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    expect(await playOneCard(page)).toBe(true);
    await expect(page.getByTestId(/^hand-card-/)).toHaveCount(6, { timeout: TIMEOUT_TRANSITION });
  });

  // 花札は手続き描画なので、共有のトランプ絵ではなく月と札種のグリフが出る。
  test('draws the hanafuda cards procedurally', async ({ page }) => {
    await navigateTo(page, '/sakura');
    const field = page.getByTestId('field-card-0');
    await expect(field).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(field.locator('svg, [data-deck="hanafuda"]').first()).toBeVisible();
  });

  test('plays the round out and reports the result', async ({ page }) => {
    await navigateTo(page, '/sakura');
    await expect(page.getByTestId('hand-card-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    for (let i = 0; i < 12; i++) {
      if (await isVisibleWithin(page.getByTestId('sakura-round-result'), 500)) break;
      if (!(await playOneCard(page))) break;
    }

    const result = page.getByTestId('sakura-round-result');
    await expect(result).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    // 席ごとの内訳が並ぶ。
    await expect(page.getByTestId('sakura-round-seat-0')).toBeVisible();

    await page.getByTestId('sakura-next-round').click();
    await waitForLoaded(page);
    // 配り直されて手札が戻る。
    await expect(page.getByTestId(/^hand-card-/)).toHaveCount(7, { timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/sakura');
    await expect(page.getByTestId('hand-card-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await playOneCard(page);

    await page.getByRole('button', { name: 'リセット' }).click();
    const confirm = page.getByRole('button', { name: 'リセットする' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByTestId(/^hand-card-/)).toHaveCount(7, { timeout: TIMEOUT_TRANSITION });
  });
});
