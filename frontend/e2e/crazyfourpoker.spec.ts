import { expect, test } from '@playwright/test';
import { isVisibleWithin, navigateTo, TIMEOUT_ACTION, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Crazy 4 Poker E2E', () => {
  test('deals, plays a hand and moves to the next one', async ({ page }) => {
    await navigateTo(page, '/crazyfourpoker');

    const deal = page.getByRole('button', { name: '配る' });
    await expect(deal).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await deal.click();
    await waitForLoaded(page);

    // 判断フェーズ。**倍率のボタンは手役次第で 1 つか 3 つ**なので、常にある 1 倍を押す。
    const play1 = page.getByTestId('c4p-play-1');
    await expect(play1).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    // **判断中はディーラーの手が出ていないこと。**
    await expect(page.getByTestId('c4p-dealer-hidden')).toBeVisible();
    await expect(page.getByTestId('c4p-dealer-hand')).toHaveCount(0);

    await play1.click();
    await waitForLoaded(page);

    // 決着するとディーラーの手が開き、次へ進めるようになる。
    await expect(page.getByTestId('c4p-dealer-hand')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByTestId('c4p-result')).toBeVisible({ timeout: TIMEOUT_ACTION });

    const next = page.getByRole('button', { name: '次のラウンド' });
    await expect(next).toBeVisible({ timeout: TIMEOUT_ACTION });
    await next.click();
    await waitForLoaded(page);

    await expect(page.getByRole('button', { name: '配る' })).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  });

  test('folding settles the hand', async ({ page }) => {
    await navigateTo(page, '/crazyfourpoker');
    await page.getByRole('button', { name: '配る' }).click();
    await waitForLoaded(page);

    const fold = page.getByRole('button', { name: '降りる' });
    await expect(fold).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await fold.click();
    await waitForLoaded(page);

    await expect(page.getByTestId('c4p-result')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: '次のラウンド' })).toBeVisible({ timeout: TIMEOUT_ACTION });
  });

  test('the 3x button only appears with a strong hand', async ({ page }) => {
    await navigateTo(page, '/crazyfourpoker');

    // **配り依存なので、出るまで賭け直す**のではなく、出ている本数と告知の整合を見る。
    await page.getByRole('button', { name: '配る' }).click();
    await waitForLoaded(page);
    await expect(page.getByTestId('c4p-play-1')).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    const strong = await isVisibleWithin(page.getByTestId('c4p-play-3'), TIMEOUT_ACTION);
    const notice = await page.getByTestId('c4p-multiplier-notice').textContent();
    if (strong) {
      // 3 倍が出ているなら 2 倍も出ていて、告知は強い手を言っているはず。
      await expect(page.getByTestId('c4p-play-2')).toBeVisible();
      expect(notice).toContain('3倍');
    } else {
      // 出ていないなら 2 倍も無く、告知はエース未満を言っているはず。
      await expect(page.getByTestId('c4p-play-2')).toHaveCount(0);
      expect(notice).toContain('エース未満');
    }
  });
});
