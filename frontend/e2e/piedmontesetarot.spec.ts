import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

/**
 * Selects the first `count` selectable cards in the human's hand.
 *
 * **`data-legal` is what marks a selectable card**, and the hand section is
 * found by `data-tutorial` — it has no `data-testid`. A locator on the wrong
 * attribute matches nothing and the test then "passes" by selecting zero cards.
 */
async function selectCards(page: Parameters<typeof navigateTo>[0], count: number): Promise<number> {
  const cards = page.locator('[data-tutorial="piedmontesetarot-player-hand"] button[data-legal]');
  const available = await cards.count();
  const picked = Math.min(count, available);
  for (let i = 0; i < picked; i++) {
    await cards.nth(i).click();
  }
  return picked;
}

test.describe('Tarocco Piemontese E2E', () => {
  test('deals a four-handed table and asks the dealer for the talon', async ({ page }) => {
    await navigateTo(page, '/piedmontesetarot');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // 席 0 が親なので、開幕はスカルト待ち。**捨てる枚数はタロンぶん (2 枚)。**
    const prompt = page.getByTestId('piedmontesetarot-discard-prompt');
    const waiting = page.getByTestId('piedmontesetarot-waiting');
    await expect(prompt.or(waiting).first()).toBeVisible({ timeout: TIMEOUT_ACTION });
    if (await prompt.isVisible()) {
      await expect(prompt).toContainText('/2');
    }
  });

  // **タロンを捨てるとプレイに入る。** ここが動かないと 1 枚も出せない。
  test('buries the talon and reaches the play phase', async ({ page }) => {
    await navigateTo(page, '/piedmontesetarot');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const discardButton = page.getByRole('button', { name: /捨てる/ });
    if (!(await isVisibleWithin(discardButton, TIMEOUT_ACTION))) return; // CPU が親の配り
    await expect(discardButton).toBeDisabled();

    const picked = await selectCards(page, 2);
    expect(picked).toBe(2);
    await expect(discardButton).toBeEnabled();
    await discardButton.click();
    await waitForLoaded(page);

    // 捨てたらスカルトの案内は消える。
    await expect(page.getByTestId('piedmontesetarot-discard-prompt')).toHaveCount(0, {
      timeout: TIMEOUT_TRANSITION,
    });
    await expect(page.getByText(/トリック 1\/19/)).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('plays a card once the tricks start', async ({ page }) => {
    await navigateTo(page, '/piedmontesetarot');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const discardButton = page.getByRole('button', { name: /捨てる/ });
    if (await isVisibleWithin(discardButton, TIMEOUT_ACTION)) {
      await selectCards(page, 2);
      await discardButton.click();
      await waitForLoaded(page);
    }

    const playButton = page.getByRole('button', { name: '出す' });
    if (!(await isVisibleWithin(playButton, TIMEOUT_ACTION))) return; // まだ CPU の手番
    await selectCards(page, 1);
    await playButton.click();
    await waitForLoaded(page);
    // 何かしら盤面が進む: 「次のトリック」か、次の自分の手番。
    const next = page.getByRole('button', { name: '次のトリック' });
    await expect(next.or(playButton).first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('resets from the footer', async ({ page }) => {
    await navigateTo(page, '/piedmontesetarot');
    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: 'リセット', exact: true }).click();
    const confirm = page.getByRole('button', { name: '確認' });
    if (await isVisibleWithin(confirm, TIMEOUT_ACTION)) await confirm.click();
    await waitForLoaded(page);

    await expect(page.getByText('ディール 1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });
});
