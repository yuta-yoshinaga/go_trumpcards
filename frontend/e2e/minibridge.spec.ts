import { expect, type Page, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION } from './helpers';

// **合法な札だけを選ぶ。** カードは意図的に disabled にしていないので、
// フォロー義務を満たさない札を押すとサーバが拒否して盤面が動かない。
const legalCard = (page: Page) => page.locator('button.ring-ds-success');

/** Your own hand — the dummy's buttons are labelled differently on purpose. */
const ownHand = (page: Page) => page.getByRole('button', { name: /^(?!ダミーの).*を出す$/ });

/** Contract buttons, only rendered when the human is the declarer. */
const contractBtn = (page: Page) => page.locator('[data-testid^="mb-contract-"]:not([disabled])');

/**
 * True when the human sits in the dummy seat.
 *
 * **落札者が両方の手を打つので、そのとき人間には一度も手番が来ない** ——
 * そしてディールは人間の入力を一切待たずに 13 トリックぶん進み切る。契約が
 * 決まった直後に見に行っても、盤面は既に「ディール終了」で全員の手札が空。
 *
 * ダミーは落札者の相方 (declarer+2)%4、落札者は HCP 最大の席なので、これは
 * 2000 配りの実測で 26.6% 起きる。押せる札や公開されたダミーの手札がある
 * ことを無条件に期待すると、その配りでは確定的に落ちる (issue #5381)。
 */
const humanIsDummy = async (page: Page) => {
  for (const id of [0, 1, 2, 3]) {
    const text = ((await page.getByTestId(`mb-seat-${id.toString()}`).textContent()) ?? '').trim();
    if (!/^(あなた|You)/.test(text)) continue;
    return /\[(ダミー|dummy)\]/.test(text);
  }
  throw new Error('no seat row is labelled as the human — the seat header changed');
};

/** Tricks taken across all four seats; 13 once the deal is played out. */
const totalTricks = async (page: Page) => {
  let total = 0;
  for (const id of [0, 1, 2, 3]) {
    const text = (await page.getByTestId(`mb-seat-${id.toString()}`).textContent()) ?? '';
    const match = text.match(/(?:獲得|tricks)\s*(\d+)/);
    expect(match, `seat ${id.toString()} shows its trick count`).not.toBeNull();
    total += Number(match?.[1] ?? 0);
  }
  return total;
};

test.describe('Minibridge E2E', () => {
  test('navigates to minibridge and renders initial game state', async ({ page }) => {
    await navigateTo(page, '/minibridge');

    await expect(page.getByText(/ミニブリッジ|Minibridge/).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('mb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **競りが無いこと自体が規則。**
  test('always states that there is no auction', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await expect(page.getByTestId('mb-rule')).toContainText(/競り|auction/, { timeout: TIMEOUT_TRANSITION });
  });

  // **HCP は 4 席ぶん公開され、合計は必ず 40。**
  test('shows every seat HCP, totalling forty', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await expect(page.getByTestId('mb-seat-0')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    let total = 0;
    for (const id of [0, 1, 2, 3]) {
      const text = (await page.getByTestId(`mb-seat-${id}`).textContent()) ?? '';
      const match = text.match(/HCP\s*(\d+)/);
      expect(match, `seat ${id} shows its HCP`).not.toBeNull();
      total += Number(match?.[1] ?? 0);
    }
    expect(total).toBe(40);
    await expect(page.getByTestId('mb-seat-4')).toHaveCount(0);
  });

  /** Settle the contract — the human only chooses it when they are the declarer. */
  const settleContract = async (page: Page) => {
    await expect(page.getByTestId('mb-round')).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await contractBtn(page).first().isVisible()) {
      await contractBtn(page).first().click();
    }
    // どちらの場合も契約が決まってプレイに入る。
    await expect(page.getByTestId('mb-contract')).not.toContainText(/未決定|not yet chosen/, {
      timeout: TIMEOUT_ACTION,
    });
  };

  test('settles the contract and reveals the dummy', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await settleContract(page);

    if (await humanIsDummy(page)) {
      // 自分がダミーなら、公開されるのは自分の手札 —— 席行のダミー表示がその印。
      // 盤面は既にディールを打ち切っているので、**素通りさせずそれを主張する**。
      await expect.poll(async () => totalTricks(page), { timeout: TIMEOUT_ACTION }).toBe(13);
      return;
    }
    // **ダミーは契約が決まると公開される。**
    await expect(page.getByTestId('mb-dummy')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('plays a card and the hand shrinks', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await settleContract(page);

    if (await humanIsDummy(page)) {
      // 人間がダミーの配りには出す手札が残っていない。**素通りさせず**、
      // CPU が最後まで打ち切ったこと（13 トリック）と、押せる札が 1 枚も
      // 出ていないことを主張する —— 出せたらサーバが拒否するのでバグ。
      await expect.poll(async () => totalTricks(page), { timeout: TIMEOUT_ACTION }).toBe(13);
      await expect(legalCard(page)).toHaveCount(0);
      return;
    }

    // **必ず動く信号は手札の枚数。** 取ったかどうかに依らない。
    const before = (await ownHand(page).count()) + (await page.getByRole('button', { name: /^ダミーの/ }).count());
    expect(before).toBeGreaterThan(0);

    await expect(legalCard(page).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    await expect(legalCard(page).first()).toBeEnabled({ timeout: TIMEOUT_ACTION });
    await legalCard(page).first().click();

    await expect
      .poll(
        async () => (await ownHand(page).count()) + (await page.getByRole('button', { name: /^ダミーの/ }).count()),
        { timeout: TIMEOUT_ACTION },
      )
      .toBeLessThan(before);
  });

  test('can reset the game via the reset confirmation dialog', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await expect(page.getByTestId('mb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^リセット$|^Reset$/ })
      .first()
      .click();
    await expect(page.getByRole('alertdialog')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await page
      .getByRole('button', { name: /^確認$|^Confirm$/ })
      .first()
      .click();
    await expect(page.getByTestId('mb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('give up ends the game', async ({ page }) => {
    await navigateTo(page, '/minibridge');
    await expect(page.getByTestId('mb-round')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page
      .getByRole('button', { name: /^投了$|^Give up$/ })
      .first()
      .click();
    await expect(page.getByTestId('mb-result')).toBeVisible({ timeout: TIMEOUT_ACTION });
  });
});
