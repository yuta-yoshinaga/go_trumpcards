import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

/** The human's hand. **The only signal that always moves** when a card is played. */
const hand = (page: Page) => page.getByRole('button', { name: /を出す|^Play / });

test.describe('Honeymoon Bridge E2E', () => {
  test('navigates to honeymoonbridge and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');

    await expect(page.getByText(/ハネムーンブリッジ|Honeymoon Bridge/).first()).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByTestId('hb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **前半のトリックは得点にならない。** 読めないと打ち方を間違える。
  test('says the first half does not score, and shows the stock', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await expect(page.getByTestId('hb-rule')).toContainText(/得点|score/, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('hb-stock')).toContainText('26', { timeout: TIMEOUT_TRANSITION });
  });

  test('labels both seats', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    for (const id of [0, 1]) {
      await expect(page.getByTestId(`hb-seat-${id}`)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    }
    // 2 人専用。3 席目は無い。
    await expect(page.getByTestId('hb-seat-2')).toHaveCount(0);
  });

  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });

    const before = await hand(page).count();
    expect(before).toBeGreaterThan(0);
    await legalCard(page).first().click();

    // 引き合いは打った直後に山札から引くので、枚数は減ってから戻る。
    // **山札は必ず 2 枚減る**ので、そちらを見る。
    await expect
      .poll(async () => (await page.getByTestId('hb-stock').textContent()) ?? '', { timeout: TIMEOUT_ACTION })
      .toContain('24');
  });

  /** Play out the 13 scoreless tricks so the auction opens. */
  const reachTheAuction = async (page: Page) => {
    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    for (let i = 0; i < 13; i++) {
      if (await page.getByTestId('hb-pass-btn').isVisible()) break;
      await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
      await legalCard(page).first().click();
      await expect(legalCard(page).first().or(page.getByTestId('hb-pass-btn'))).toBeVisible({
        timeout: TIMEOUT_ACTION,
      });
    }
    await expect(page.getByTestId('hb-pass-btn')).toBeVisible({ timeout: TIMEOUT_ACTION });
  };

  // **山札を使い切ると競りに入る。** 26 枚配り切って両者 13 枚に戻る。
  test('the draw phase empties the stock and opens the auction', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await reachTheAuction(page);

    await expect(page.getByTestId('hb-stock')).toHaveCount(0);
    await expect(page.getByTestId('hb-minbid')).toBeVisible({ timeout: TIMEOUT_ACTION });
    // 両者 13 枚に戻っている。
    await expect.poll(async () => hand(page).count(), { timeout: TIMEOUT_ACTION }).toBe(13);
  });

  // **押せるボタンはサーバが受理する。** 拒否される値を出していれば盤面が動かない。
  test('bidding the offered contract is accepted', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await reachTheAuction(page);

    const bid = page.locator('[data-testid^="hb-bid-"]:not([disabled])').first();
    await expect(bid).toBeVisible({ timeout: TIMEOUT_ACTION });
    await bid.click();

    // 契約が決まる（相手が上を宣言しても「未決定」ではなくなる）。
    await expect(page.getByTestId('hb-contract')).not.toContainText(/未決定|not yet decided/, {
      timeout: TIMEOUT_ACTION,
    });
  });

  test('passing is accepted', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await reachTheAuction(page);

    // **1回パスしても競りは終わらない。** 締まるのは連続2回パスしたときだけで
    // (domain の passCount)、相手が上を宣言すると passCount は 0 に戻り、手番は
    // こちらへ返ってくる —— つまりパスボタンは再び現れる。「パスしたらボタンが
    // 消える」と書いていたので、CPU が宣言した配りでだけ落ちていた。
    const contractBefore = await page.getByTestId('hb-contract').innerText();
    await page.getByTestId('hb-pass-btn').click();

    // パスが受理されたなら、競りが締まる (ボタンが消える) か、相手が上を宣言して
    // 契約表示が変わるか、どちらかは必ず起きる。どちらも起きないのは拒否された時。
    await expect
      .poll(
        async () => {
          if ((await page.getByTestId('hb-pass-btn').count()) === 0) return true;
          return (await page.getByTestId('hb-contract').innerText()) !== contractBefore;
        },
        { timeout: TIMEOUT_ACTION },
      )
      .toBe(true);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await expect(page.getByTestId('hb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('hb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/honeymoonbridge');
    await expect(page.getByTestId('hb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('hb-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
