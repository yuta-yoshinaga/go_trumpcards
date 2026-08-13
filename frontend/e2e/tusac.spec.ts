import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/** Plays one full turn: draw, then discard the first card. */
async function playTurn(page: Parameters<typeof navigateTo>[0]) {
  const draw = page.getByTestId('tusac-draw');
  if (!(await isVisibleWithin(draw, TIMEOUT_ACTION))) return false;
  await draw.click();
  await waitForLoaded(page);

  const card = page.getByTestId('tusac-card-0');
  if (!(await isVisibleWithin(card, TIMEOUT_ACTION))) return false;
  await card.click();
  const discard = page.getByTestId('tusac-discard');
  if (!(await isVisibleWithin(discard, TIMEOUT_ACTION))) return false;
  await discard.click();
  await waitForLoaded(page);
  return true;
}

test.describe('Tu Sac E2E', () => {
  test('deals the four-colour deck and shows only your own hand', async ({ page }) => {
    await navigateTo(page, '/tusac');
    await expect(page.getByTestId('tusac-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **自分の手札は 20 枚。**
    await expect(page.getByTestId('tusac-hand').locator('> *')).toHaveCount(20);

    // **相手の手札は届かない。** 枚数だけが出る。
    await expect(page.getByTestId('tusac-count-1')).toContainText('20');
    await expect(page.getByTestId('tusac-melds-1')).toBeVisible();

    // 山と捨て札が始まっている。
    await expect(page.getByTestId('tusac-stock')).toBeVisible();
    await expect(page.getByTestId('tusac-discard-top')).toBeVisible();
  });

  // **引く場面と出す場面でボタンが入れ替わる。** 同時には出ない。
  test('asks to draw, then to meld or discard', async ({ page }) => {
    await navigateTo(page, '/tusac');
    await expect(page.getByTestId('tusac-draw')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('tusac-take')).toBeVisible();
    await expect(page.getByTestId('tusac-meld')).toHaveCount(0);
    await expect(page.getByTestId('tusac-discard')).toHaveCount(0);

    await page.getByTestId('tusac-draw').click();
    await waitForLoaded(page);

    await expect(page.getByTestId('tusac-meld')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('tusac-discard')).toBeVisible();
    await expect(page.getByTestId('tusac-draw')).toHaveCount(0);

    // 何も選んでいなければ押せない。
    await expect(page.getByTestId('tusac-meld')).toBeDisabled();
    await expect(page.getByTestId('tusac-discard')).toBeDisabled();
  });

  // **選んだ札だけが動く。** 引いて捨てると手札は 20 枚に戻る。
  test('draws and discards, keeping the hand at twenty', async ({ page }) => {
    await navigateTo(page, '/tusac');
    await expect(page.getByTestId('tusac-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByTestId('tusac-draw').click();
    await waitForLoaded(page);
    await expect(page.getByTestId('tusac-hand').locator('> *')).toHaveCount(21);

    await page.getByTestId('tusac-card-0').click();
    await expect(page.getByTestId('tusac-selected')).toContainText('1');
    await page.getByTestId('tusac-discard').click();
    await waitForLoaded(page);

    await expect(page.getByTestId('tusac-hand').locator('> *')).toHaveCount(20, {
      timeout: TIMEOUT_TRANSITION,
    });
    // 手番が戻り、また引く場面になる。
    await expect(page.getByTestId('tusac-draw')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('plays turns until the round ends', async ({ page }) => {
    await navigateTo(page, '/tusac');
    await expect(page.getByTestId('tusac-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    for (let turn = 0; turn < 12; turn++) {
      if (await isVisibleWithin(page.getByTestId('tusac-next'), TIMEOUT_ACTION)) break;
      if (!(await playTurn(page))) break;
    }

    // 山が尽きるか上がりでラウンドが終わる。
    const next = page.getByTestId('tusac-next');
    if (await isVisibleWithin(next, TIMEOUT_ACTION)) {
      await next.click();
      await waitForLoaded(page);
      await expect(page.getByTestId('tusac-round')).toContainText('2', { timeout: TIMEOUT_TRANSITION });
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/tusac');
    await expect(page.getByTestId('tusac-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /リセット|やり直/ })
      .first()
      .click();
    const confirm = page.getByRole('button', { name: /はい|OK|確認/ });
    if (await isVisibleWithin(confirm.first(), TIMEOUT_ACTION)) {
      await confirm.first().click();
    }
    await waitForLoaded(page);
    await expect(page.getByTestId('tusac-round')).toContainText('1', { timeout: TIMEOUT_TRANSITION });
  });
});
