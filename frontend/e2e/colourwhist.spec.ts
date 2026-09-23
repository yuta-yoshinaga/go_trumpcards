import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

test.describe('Colour Whist E2E', () => {
  test('reaches a contract and plays a card', async ({ page }) => {
    await navigateTo(page, '/colourwhist');

    // **配りによって始まる場面が違います。** トルールが出れば競りは飛ばされます。
    //
    // 契約ボタンは**見えていても押せない**ことがあります（CPU が先に同じ契約まで
    // 競り上げていると disabled になる）。`isVisible()` は disabled でも true を
    // 返すので、そのまま click するとタイムアウトします。常に押せる「降りる」で
    // 競りを進めます。
    const pass = page.getByRole('button', { name: '降りる' });
    if (await pass.isVisible()) {
      await pass.click();
      await waitForLoaded(page);
    }

    // 切り札を選ぶ場面が来たら選ぶ（押せる場合だけ）。
    const heart = page.locator('button:not([disabled])', { hasText: /ハート.*切り札/ });
    if (await heart.first().isVisible()) {
      await heart.first().click();
      await waitForLoaded(page);
    }

    await expect(page.getByTestId('colourwhist-contract')).toBeVisible({ timeout: TIMEOUT_ACTION });

    // **出せる札があるときだけ出します。**
    //
    // 出せる札が無い理由は 2 つあり、区別しないと嘘の失敗になります。ページは
    // `playable = isPlayPhase && state.isHumanTurn && legal.has(i)` で
    // `aria-disabled` を決める (`ColourWhistPage.tsx:267`) ので、
    // **CPU の手番のあいだは手札が満杯でも全部 disabled** になります。
    // 「押せる札が無い == ラウンドが終わった」と読むと、CPU が先導している
    // ふつうの局面で `cw-next-button` を待って落ちます (配り依存)。
    const handButtons = page.locator('[data-testid="colourwhist-hand"] button');
    const legal = handButtons.locator(':scope:not([aria-disabled="true"])');
    if (await legal.first().isVisible()) {
      const before = await handButtons.count();
      await legal.first().click();
      await waitForLoaded(page);
      await expect(handButtons).toHaveCount(before - 1, { timeout: TIMEOUT_ACTION });
    } else if ((await handButtons.count()) === 0) {
      // 手札そのものが無いならラウンドが終わっている——次へ進めることを確かめます。
      await expect(page.getByTestId('cw-next-button')).toBeVisible({ timeout: TIMEOUT_ACTION });
    } else {
      // 札はあるが自分の手番でない。盤が生きていることだけ確かめて終わります。
      await expect(page.getByTestId('colourwhist-seats')).toBeVisible({ timeout: TIMEOUT_ACTION });
    }
  });

  test('shows the contract and seats on load', async ({ page }) => {
    await navigateTo(page, '/colourwhist');

    await expect(page.getByTestId('colourwhist-contract')).toBeVisible();
    await expect(page.getByTestId('colourwhist-seats')).toBeVisible();
    // トルールが出た配りなら、競りが飛ばされた理由が表示される。
    const notice = page.getByTestId('colourwhist-troel-notice');
    if (await notice.isVisible()) {
      await expect(notice).toContainText('エース3枚');
    }
  });
});
