import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Selects the first selectable card in the human's hand and reports whether one
 * was found.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — it carries no `data-testid`. A locator on the
 * wrong attribute matches nothing, and the test then "passes" having clicked
 * nothing at all.
 */
async function selectFirstCard(page: Parameters<typeof navigateTo>[0]): Promise<boolean> {
  const cards = page.locator('[data-tutorial="unsunkaruta-player-hand"] button[data-legal]');
  if ((await cards.count()) === 0) return false;
  await cards.first().click();
  return true;
}

test.describe('Unsun Karuta E2E', () => {
  test('deals eight seats with a turned trump', async ({ page }) => {
    await navigateTo(page, '/unsunkaruta');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByText('トリック 1/9')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **切り札は返した 1 枚で決まる。** 数札の強弱がスートで逆になるので、
    // どのスートが切り札かが出ていないと盤面が読めない。
    await expect(page.getByTestId('unsunkaruta-trump')).toBeVisible();
    await expect(page.getByTestId('unsunkaruta-teams')).toContainText('味方');
  });

  test('plays a card and takes the trick forward', async ({ page }) => {
    await navigateTo(page, '/unsunkaruta');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const playButton = page.getByTestId('unsunkaruta-play');
    if (!(await isVisibleWithin(playButton, TIMEOUT_ACTION))) return; // まだ CPU の手番
    await expect(playButton).toBeDisabled();

    expect(await selectFirstCard(page)).toBe(true);
    await expect(playButton).toBeEnabled();
    await playButton.click();
    await waitForLoaded(page);

    // 盤面は進む: トリックが決まったか、次の自分の手番か。
    const next = page.getByTestId('unsunkaruta-next-trick');
    await expect(next.or(playButton).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  // **宣言はリードのときだけ。** 席 0 は第 1 ディールの親なので、リードは
  // 必ず CPU 1 から始まる ── ここで declare ボタンが出ていたら条件が壊れている。
  // トリックを取ればリードが回ってくるので、出てきたところで実際に押す。
  test('offers meri only on a lead, and declaring from one advances the board', async ({ page }) => {
    await navigateTo(page, '/unsunkaruta');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const play = page.getByTestId('unsunkaruta-play');
    const declare = page.getByTestId('unsunkaruta-declare');
    await expect(play).toBeVisible({ timeout: TIMEOUT_ACTION });
    // 開幕は親の左隣 (CPU 1) がリード。宣言の面はここには無い。
    await expect(declare).toHaveCount(0);
    await expect(page.getByTestId('unsunkaruta-can-declare')).toHaveCount(0);

    // 取れば次のトリックはこちらのリード。9 トリック回して出てきたら押す。
    let declared = false;
    for (let trick = 0; trick < 9 && !declared; trick++) {
      const next = page.getByTestId('unsunkaruta-next-trick');
      if (await next.isVisible()) {
        await next.click();
        await waitForLoaded(page);
      }
      if (!(await play.isVisible())) break;
      if (await declare.isVisible()) {
        await expect(page.getByTestId('unsunkaruta-can-declare')).toBeVisible();
        expect(await selectFirstCard(page)).toBe(true);
        await declare.click();
        declared = true;
      } else {
        expect(await selectFirstCard(page)).toBe(true);
        await play.click();
      }
      await waitForLoaded(page);
    }

    // リードが一度も回ってこないディールもある。押せたときだけ、盤面が進んだ
    // ことまで見る。
    if (declared) {
      await expect(page.getByTestId('unsunkaruta-next-trick').or(play).first()).toBeVisible({
        timeout: TIMEOUT_TRANSITION,
      });
    }
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/unsunkaruta');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
