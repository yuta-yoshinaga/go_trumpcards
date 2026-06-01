import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Macau E2E', () => {
  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/macau');

    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const playButton = page.getByRole('button', { name: '出す' });
    const drawButton = page.getByRole('button', { name: '引く' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const declareButton = page.getByRole('button', { name: 'マカオ！' });
    const skipDeclareButton = page.getByRole('button', { name: '宣言しない' });
    const suitSpade = page.getByRole('button', { name: '♠ スペード' });
    const endResetButton = page.getByRole('button', { name: '次のゲーム' });
    const takePenaltyButton = page.getByRole('button', { name: /^\d+枚引き受ける$/ });
    const handCards = page.locator('button[aria-pressed]:has(img)');

    const MAX_TURNS = 100;
    let interactions = 0;
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      if (await endResetButton.isVisible()) break;

      await expect(
        playButton.or(drawButton).or(nextRoundButton).or(declareButton).or(suitSpade).or(endResetButton).first(),
      ).toBeVisible({ timeout: 10_000 });

      // Declaration phase: declare "Macau!"
      if (await declareButton.isVisible()) {
        await declareButton.click();
        await waitForLoaded(page);
        interactions++;
        continue;
      }

      // Suit choice phase
      if (await suitSpade.isVisible()) {
        await suitSpade.click();
        await waitForLoaded(page);
        interactions++;
        continue;
      }

      const playVisible = await playButton.isVisible();
      const drawVisible = await drawButton.isVisible();
      const takeVisible = await takePenaltyButton.isVisible();

      if (playVisible || drawVisible || takeVisible) {
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible()) && (await playButton.isEnabled())) {
          await playButton.click();
          await waitForLoaded(page);
          continue;
        }
        if ((await takePenaltyButton.isVisible()) && (await takePenaltyButton.isEnabled())) {
          await takePenaltyButton.click();
          await waitForLoaded(page);
          continue;
        }
        if ((await drawButton.isVisible()) && (await drawButton.isEnabled())) {
          await drawButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      if (await nextRoundButton.isVisible()) {
        await nextRoundButton.click();
        await waitForLoaded(page);
        interactions++;
        continue;
      }

      // Fallback: if skip-declare is somehow the only option
      if (await skipDeclareButton.isVisible()) {
        await skipDeclareButton.click();
        await waitForLoaded(page);
        interactions++;
      }
    }

    expect(interactions).toBeGreaterThan(0);

    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
